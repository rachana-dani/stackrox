package sensor

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/stackrox/rox/generated/internalapi/central"
	sensorInternal "github.com/stackrox/rox/generated/internalapi/sensor"
	"github.com/stackrox/rox/generated/storage"
	"github.com/stackrox/rox/pkg/centralsensor"
	"github.com/stackrox/rox/pkg/clusterid"
	"github.com/stackrox/rox/pkg/env"
	"github.com/stackrox/rox/pkg/expiringcache"
	"github.com/stackrox/rox/pkg/grpc"
	"github.com/stackrox/rox/pkg/logging"
	"github.com/stackrox/rox/pkg/namespaces"
	"github.com/stackrox/rox/pkg/protoutils"
	"github.com/stackrox/rox/pkg/satoken"
	"github.com/stackrox/rox/sensor/common"
	"github.com/stackrox/rox/sensor/common/admissioncontroller"
	"github.com/stackrox/rox/sensor/common/certdistribution"
	"github.com/stackrox/rox/sensor/common/clusterentities"
	"github.com/stackrox/rox/sensor/common/compliance"
	"github.com/stackrox/rox/sensor/common/config"
	"github.com/stackrox/rox/sensor/common/deployment"
	"github.com/stackrox/rox/sensor/common/detector"
	"github.com/stackrox/rox/sensor/common/externalsrcs"
	"github.com/stackrox/rox/sensor/common/image"
	"github.com/stackrox/rox/sensor/common/networkflow/manager"
	"github.com/stackrox/rox/sensor/common/networkflow/service"
	"github.com/stackrox/rox/sensor/common/processfilter"
	"github.com/stackrox/rox/sensor/common/processsignal"
	"github.com/stackrox/rox/sensor/common/registry"
	"github.com/stackrox/rox/sensor/common/reprocessor"
	"github.com/stackrox/rox/sensor/common/sensor"
	"github.com/stackrox/rox/sensor/common/sensor/helmconfig"
	signalService "github.com/stackrox/rox/sensor/common/signal"
	k8sadmctrl "github.com/stackrox/rox/sensor/kubernetes/admissioncontroller"
	"github.com/stackrox/rox/sensor/kubernetes/clusterhealth"
	"github.com/stackrox/rox/sensor/kubernetes/clustermetrics"
	"github.com/stackrox/rox/sensor/kubernetes/clusterstatus"
	"github.com/stackrox/rox/sensor/kubernetes/enforcer"
	"github.com/stackrox/rox/sensor/kubernetes/listener"
	"github.com/stackrox/rox/sensor/kubernetes/listener/resources"
	"github.com/stackrox/rox/sensor/kubernetes/localscanner"
	"github.com/stackrox/rox/sensor/kubernetes/networkpolicies"
	"github.com/stackrox/rox/sensor/kubernetes/orchestrator"
	"github.com/stackrox/rox/sensor/kubernetes/telemetry"
	"github.com/stackrox/rox/sensor/kubernetes/upgrade"
)

var (
	log = logging.LoggerForModule()
)

// CreateSensor takes in a client interface and returns a sensor instantiation
func CreateSensor(cfg *CreateOptions) (*sensor.Sensor, error) {
	admCtrlSettingsMgr := admissioncontroller.NewSettingsManager(resources.DeploymentStoreSingleton(), resources.PodStoreSingleton())

	var helmManagedConfig *central.HelmManagedConfigInit
	if configFP := helmconfig.HelmConfigFingerprint.Setting(); configFP != "" {
		var err error
		helmManagedConfig, err = helmconfig.Load()
		if err != nil {
			return nil, errors.Wrap(err, "loading Helm cluster config")
		}
		if helmManagedConfig.GetClusterConfig().GetConfigFingerprint() != configFP {
			return nil, errors.Errorf("fingerprint %q of loaded config does not match expected fingerprint %q, config changes can only be applied via 'helm upgrade' or a similar chart-based mechanism", helmManagedConfig.GetClusterConfig().GetConfigFingerprint(), configFP)
		}
		log.Infof("Loaded Helm cluster configuration with fingerprint %q", configFP)

		if err := helmconfig.CheckEffectiveClusterName(helmManagedConfig); err != nil {
			return nil, errors.Wrap(err, "validating cluster name")
		}
	}

	if helmManagedConfig.GetClusterName() == "" {
		certClusterID, err := clusterid.ParseClusterIDFromServiceCert(storage.ServiceType_SENSOR_SERVICE)
		if err != nil {
			return nil, errors.Wrap(err, "parsing cluster ID from service certificate")
		}
		if centralsensor.IsInitCertClusterID(certClusterID) {
			return nil, errors.New("a sensor that uses certificates from an init bundle must have a cluster name specified")
		}
	}

	deploymentIdentification := fetchDeploymentIdentification(context.Background(), cfg.k8sClient.Kubernetes())
	log.Infof("Determined deployment identification: %s", protoutils.NewWrapper(deploymentIdentification))

	auditLogEventsInput := make(chan *sensorInternal.AuditEvents)
	auditLogCollectionManager := compliance.NewAuditLogCollectionManager()

	o := orchestrator.New(cfg.k8sClient.Kubernetes())
	complianceService := compliance.NewService(o, auditLogEventsInput, auditLogCollectionManager)

	configHandler := config.NewCommandHandler(admCtrlSettingsMgr, deploymentIdentification, helmManagedConfig, auditLogCollectionManager)
	enforcer, err := enforcer.New(cfg.k8sClient)
	if err != nil {
		return nil, errors.Wrap(err, "creating enforcer")
	}

	imageCache := expiringcache.NewExpiringCache(env.ReprocessInterval.DurationSetting())
	policyDetector := detector.New(enforcer, admCtrlSettingsMgr, resources.DeploymentStoreSingleton(), resources.ServiceAccountStoreSingleton(), imageCache, auditLogEventsInput, auditLogCollectionManager, resources.NetworkPolicySingleton())
	resourceListener := listener.New(cfg.k8sClient, configHandler, policyDetector, k8sNodeName.Setting(), cfg.resyncPeriod, cfg.traceWriter)
	admCtrlMsgForwarder := admissioncontroller.NewAdmCtrlMsgForwarder(admCtrlSettingsMgr, resourceListener)

	imageService := image.NewService(imageCache, registry.Singleton())
	complianceCommandHandler := compliance.NewCommandHandler(complianceService)

	// Create Process Pipeline
	indicators := make(chan *central.MsgFromSensor)
	processPipeline := processsignal.NewProcessPipeline(indicators, clusterentities.StoreInstance(), processfilter.Singleton(), policyDetector)
	processSignals := signalService.New(processPipeline, indicators)
	networkFlowManager :=
		manager.NewManager(clusterentities.StoreInstance(), externalsrcs.StoreInstance(), policyDetector)
	components := []common.SensorComponent{
		admCtrlMsgForwarder,
		enforcer,
		networkFlowManager,
		networkpolicies.NewCommandHandler(cfg.k8sClient.Kubernetes()),
		clusterstatus.NewUpdater(cfg.k8sClient),
		clusterhealth.NewUpdater(cfg.k8sClient.Kubernetes(), 0),
		clustermetrics.New(cfg.k8sClient.Kubernetes()),
		complianceCommandHandler,
		processSignals,
		telemetry.NewCommandHandler(cfg.k8sClient.Kubernetes()),
		externalsrcs.Singleton(),
		admissioncontroller.AlertHandlerSingleton(),
		auditLogCollectionManager,
		reprocessor.NewHandler(admCtrlSettingsMgr, policyDetector, imageCache),
	}

	if !cfg.localSensor {
		upgradeCmdHandler, err := upgrade.NewCommandHandler(configHandler)
		if err != nil {
			return nil, errors.Wrap(err, "creating upgrade command handler")
		}
		components = append(components, upgradeCmdHandler)
	}

	sensorNamespace, err := satoken.LoadNamespaceFromFile()
	if err != nil {
		log.Errorf("Failed to determine namespace from service account token file: %s", err)
	}
	if sensorNamespace == "" {
		sensorNamespace = os.Getenv("POD_NAMESPACE")
	}
	if sensorNamespace == "" {
		sensorNamespace = namespaces.StackRox
		log.Warnf("Unable to determine Sensor namespace, defaulting to %s", sensorNamespace)
	}

	if admCtrlSettingsMgr != nil {
		components = append(components, k8sadmctrl.NewConfigMapSettingsPersister(cfg.k8sClient.Kubernetes(), admCtrlSettingsMgr, sensorNamespace))
	}

	// Local scanner can be started even if scanner-tls certs are available in the same namespace because
	// it ignores secrets not owned by Sensor.
	if securedClusterIsNotManagedManually(helmManagedConfig) && env.LocalImageScanningEnabled.BooleanSetting() {
		podName := os.Getenv("POD_NAME")
		components = append(components,
			localscanner.NewLocalScannerTLSIssuer(cfg.k8sClient.Kubernetes(), sensorNamespace, podName))
	}

	s := sensor.NewSensor(
		configHandler,
		policyDetector,
		imageService,
		cfg.centralConnFactory,
		components...,
	)

	if cfg.workloadManager != nil {
		cfg.workloadManager.SetSignalHandlers(processPipeline, networkFlowManager)
	}

	networkFlowService := service.NewService(networkFlowManager)
	apiServices := []grpc.APIService{
		networkFlowService,
		processSignals,
		complianceService,
		imageService,
		deployment.NewService(resources.DeploymentStoreSingleton(), resources.PodStoreSingleton()),
	}

	if admCtrlSettingsMgr != nil {
		apiServices = append(apiServices, admissioncontroller.NewManagementService(admCtrlSettingsMgr, admissioncontroller.AlertHandlerSingleton()))
	}

	apiServices = append(apiServices, certdistribution.NewService(cfg.k8sClient.Kubernetes(), sensorNamespace))

	s.AddAPIServices(apiServices...)
	return s, nil
}

func securedClusterIsNotManagedManually(helmManagedConfig *central.HelmManagedConfigInit) bool {
	return helmManagedConfig.GetManagedBy() != storage.ManagerType_MANAGER_TYPE_UNKNOWN &&
		helmManagedConfig.GetManagedBy() != storage.ManagerType_MANAGER_TYPE_MANUAL
}

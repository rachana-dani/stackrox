import React, { Component } from 'react';
import PropTypes from 'prop-types';
import { connect } from 'react-redux';
import { createStructuredSelector } from 'reselect';
import PageHeader from 'Components/PageHeader';

import integrationsList from 'Containers/Integrations/integrationsList';
import IntegrationModal from 'Containers/Integrations/IntegrationModal';
import IntegrationTile from 'Containers/Integrations/IntegrationTile';
import { actions as authActions } from 'reducers/auth';
import { actions as apiTokenActions } from 'reducers/apitokens';
import { actions as integrationActions } from 'reducers/integrations';
import { selectors } from 'reducers';
import { isBackendFeatureFlagEnabled, knownBackendFlags } from 'utils/featureFlags';
import APITokensModal from './APITokens/APITokensModal';

class IntegrationsPage extends Component {
    static propTypes = {
        authPlugins: PropTypes.arrayOf(
            PropTypes.shape({
                endpoint: PropTypes.string.isRequired,
            })
        ).isRequired,
        authProviders: PropTypes.arrayOf(
            PropTypes.shape({
                name: PropTypes.string.isRequired,
            })
        ).isRequired,
        apiTokens: PropTypes.arrayOf(
            PropTypes.shape({
                name: PropTypes.string.isRequired,
                role: PropTypes.string.isRequired,
            })
        ).isRequired,
        backups: PropTypes.arrayOf(
            PropTypes.shape({
                name: PropTypes.string.isRequired,
            })
        ).isRequired,
        notifiers: PropTypes.arrayOf(PropTypes.object).isRequired,
        imageIntegrations: PropTypes.arrayOf(PropTypes.object).isRequired,
        fetchAuthPlugins: PropTypes.func.isRequired,
        fetchAuthProviders: PropTypes.func.isRequired,
        fetchAPITokens: PropTypes.func.isRequired,
        fetchBackups: PropTypes.func.isRequired,
        fetchNotifiers: PropTypes.func.isRequired,
        fetchImageIntegrations: PropTypes.func.isRequired,
        featureFlags: PropTypes.arrayOf(
            PropTypes.shape({
                envVar: PropTypes.string.isRequired,
                enabled: PropTypes.bool.isRequired,
            })
        ).isRequired,
    };

    constructor(props) {
        super(props);

        this.state = {
            modalOpen: false,
            selectedSource: '',
            selectedType: '',
            selectedLabel: '',
        };
    }

    getEntities = (source, type) => {
        switch (source) {
            case 'authPlugins':
                this.props.fetchAuthPlugins();
                break;
            case 'authProviders':
                if (type === 'apitoken') {
                    this.props.fetchAPITokens();
                    break;
                }
                this.props.fetchAuthProviders();
                break;
            case 'imageIntegrations':
                this.props.fetchImageIntegrations();
                break;
            case 'notifiers':
                this.props.fetchNotifiers();
                break;
            case 'backups':
                this.props.fetchBackups();
                break;
            default:
                throw new Error(`Unknown source ${source}`);
        }
    };

    openIntegrationModal = (integrationCategory) => {
        this.setState({
            modalOpen: true,
            selectedSource: integrationCategory.source,
            selectedType: integrationCategory.type,
            selectedLabel: integrationCategory.label,
        });
    };

    fetchEntitiesAndCloseModal = () => {
        this.getEntities(this.state.selectedSource, this.state.selectedType);
        this.setState({
            modalOpen: false,
            selectedSource: '',
            selectedType: '',
            selectedLabel: '',
        });
    };

    findIntegrations = (source, type) => {
        const typeLowerMatches = (integration) =>
            integration.type.toLowerCase() === type.toLowerCase();

        switch (source) {
            case 'authPlugins':
                return this.props.authPlugins;
            case 'authProviders':
                if (type === 'apitoken') {
                    return this.props.apiTokens;
                }
                return this.props.authProviders.filter(typeLowerMatches);
            case 'notifiers':
                return this.props.notifiers.filter(typeLowerMatches);
            case 'backups':
                return this.props.backups.filter(typeLowerMatches);
            case 'imageIntegrations':
                return this.props.imageIntegrations.filter(typeLowerMatches);
            default:
                throw new Error(`Unknown source ${source}`);
        }
    };

    renderAPITokensModal() {
        return (
            <APITokensModal
                tokens={this.props.apiTokens}
                onRequestClose={this.fetchEntitiesAndCloseModal}
            />
        );
    }

    renderIntegrationModal() {
        const { modalOpen, selectedSource, selectedType, selectedLabel } = this.state;
        if (!modalOpen) {
            return null;
        }

        if (selectedSource === 'authProviders' && selectedType === 'apitoken') {
            return this.renderAPITokensModal();
        }

        const integrations = this.findIntegrations(selectedSource, selectedType);
        return (
            <IntegrationModal
                integrations={integrations}
                source={selectedSource}
                type={selectedType}
                label={selectedLabel}
                onRequestClose={this.fetchEntitiesAndCloseModal}
            />
        );
    }

    renderIntegrationTiles = (source) =>
        integrationsList[source].map((tile) => {
            if (tile.featureFlagDependency) {
                if (
                    isBackendFeatureFlagEnabled(
                        this.props.featureFlags,
                        tile.featureFlagDependency.featureFlag,
                        tile.featureFlagDependency.defaultValue
                    ) !== tile.featureFlagDependency.showIfValueIs
                ) {
                    return null;
                }
            }
            // TODO: remove this manual check after ROOX_HOST_SCANNING feature flag turned on
            if (
                tile.label === 'StackRox Scanner' &&
                !isBackendFeatureFlagEnabled(
                    this.props.featureFlags,
                    knownBackendFlags.ROX_HOST_SCANNING,
                    false
                )
            ) {
                // eslint-disable-next-line no-param-reassign
                tile.categories = 'Scanner';
            }
            // end TODO block to remove

            return (
                <IntegrationTile
                    key={tile.label}
                    integration={tile}
                    onClick={this.openIntegrationModal}
                    numIntegrations={this.findIntegrations(tile.source, tile.type).length}
                />
            );
        });

    render() {
        const imageIntegrations = this.renderIntegrationTiles('imageIntegrations');
        const plugins = this.renderIntegrationTiles('plugins');
        const authPlugins = this.renderIntegrationTiles('authPlugins');
        const authProviders = this.renderIntegrationTiles('authProviders');
        const backups = this.renderIntegrationTiles('backups');

        return (
            <div className="h-full flex flex-col md:w-full bg-base-200" id="integrationsPage">
                <div className="flex flex-shrink-0">
                    <PageHeader header="Integrations" subHeader="Setup & Configuration" />
                </div>
                <div className="w-full h-full overflow-auto">
                    <section className="mb-6" id="image-integrations">
                        <h2 className="bg-base-200 border-b border-primary-400 font-700 mx-4 top-0 px-3 py-4 sticky text-base text-base-600 tracking-wide  uppercase z-1">
                            Image Integrations
                        </h2>
                        <div className="flex flex-col items-center w-full">
                            <div className="flex flex-wrap w-full -mx-6 p-3">
                                {imageIntegrations}
                            </div>
                        </div>
                    </section>

                    <section className="mb-6" id="notifier-integrations">
                        <h2 className="bg-base-200 border-b border-primary-400 font-700 mx-4 top-0 px-3 py-4 sticky text-base text-base-600 tracking-wide  uppercase z-1">
                            Notifier Integrations
                        </h2>
                        <div className="flex flex-col items-center w-full">
                            <div className="flex flex-wrap w-full -mx-6 p-3">{plugins}</div>
                        </div>
                    </section>

                    <section className="mb-6" id="backup-integrations">
                        <h2 className="bg-base-200 border-b border-primary-400 font-700 mx-4 top-0 px-3 py-4 sticky text-base text-base-600 tracking-wide  uppercase z-1">
                            Backup Integrations
                        </h2>
                        <div className="flex flex-col items-center w-full">
                            <div className="flex flex-wrap w-full -mx-6 p-3">{backups}</div>
                        </div>
                    </section>

                    <section className="mb-6" id="token-integrations">
                        <h2 className="bg-base-200 border-b border-primary-400 font-700 mx-4 top-0 px-3 py-4 sticky text-base text-base-600 tracking-wide  uppercase z-1">
                            Authentication Tokens
                        </h2>
                        <div className="flex flex-col items-center w-full">
                            <div className="flex flex-wrap w-full -mx-6 p-3">{authProviders}</div>
                        </div>
                    </section>

                    <section className="mb-6" id="auth-integrations">
                        <h2 className="bg-base-200 border-b border-primary-400 font-700 mx-4 top-0 px-3 py-4 sticky text-base text-base-600 tracking-wide  uppercase z-1">
                            Authorization Plugins
                        </h2>
                        <div className="flex flex-col items-center w-full">
                            <div className="flex flex-wrap w-full -mx-6 p-3">{authPlugins}</div>
                        </div>
                    </section>
                </div>
                {this.renderIntegrationModal()}
            </div>
        );
    }
}

const mapStateToProps = createStructuredSelector({
    authPlugins: selectors.getAuthPlugins,
    authProviders: selectors.getAuthProviders,
    apiTokens: selectors.getAPITokens,
    notifiers: selectors.getNotifiers,
    imageIntegrations: selectors.getImageIntegrations,
    backups: selectors.getBackups,
    featureFlags: selectors.getFeatureFlags,
});

const mapDispatchToProps = (dispatch) => ({
    fetchAuthProviders: () => dispatch(authActions.fetchAuthProviders.request()),
    fetchAPITokens: () => dispatch(apiTokenActions.fetchAPITokens.request()),
    fetchBackups: () => dispatch(integrationActions.fetchBackups.request()),
    fetchNotifiers: () => dispatch(integrationActions.fetchNotifiers.request()),
    fetchImageIntegrations: () => dispatch(integrationActions.fetchImageIntegrations.request()),
    fetchRegistries: () => dispatch(integrationActions.fetchRegistries.request()),
    fetchScanners: () => dispatch(integrationActions.fetchScanners.request()),
    fetchAuthPlugins: () => dispatch(integrationActions.fetchAuthPlugins.request()),
});

export default connect(mapStateToProps, mapDispatchToProps)(IntegrationsPage);

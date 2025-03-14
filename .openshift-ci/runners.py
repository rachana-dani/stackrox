#!/usr/bin/env python3

"""
Common test run patterns
"""

from datetime import datetime
from clusters import NullCluster
from pre_tests import NullPreTest
from ci_tests import NullTest
from post_tests import NullPostTest


class ClusterTestSetsRunner:
    """A cluster test runner that runs multiple sets of pre, test & post steps
    wrapped by a cluster provision and with similar semantics to
    ClusterTestRunner. Each test set will attempt to run regardless of the outcome of
    prior sets. This can be overriden on a set by set basis with 'always_run'"""

    def __init__(
        self,
        cluster=NullCluster(),
        final_post=NullPostTest(),
        sets=None,
    ):
        self.cluster = cluster
        self.final_post = final_post
        if sets is None:
            sets = []
        self.sets = sets

    def run(self):
        hold = None
        try:
            self.log_event("About to provision")
            self.cluster.provision()
            self.log_event("provisioned")
        except Exception as err:
            self.log_event("ERROR: provision failed")
            hold = err

        if hold is None:
            for idx, test_set in enumerate(self.sets):
                test_set = {
                    **{
                        "name": f"set {idx + 1}",
                        "pre_test": NullPreTest(),
                        "test": NullTest(),
                        "post_test": NullPostTest(),
                        "always_run": True,
                    },
                    **test_set,
                }
                if hold is None or test_set["always_run"]:
                    try:
                        self.log_event("About to run", test_set)
                        self.run_test_set(test_set)
                        self.log_event("run completed", test_set)
                    except Exception as err:
                        self.log_event("ERROR: run failed", test_set)
                        if hold is None:
                            hold = err

        try:
            self.log_event("About to teardown")
            self.cluster.teardown()
            self.log_event("teardown completed")
        except Exception as err:
            self.log_event("ERROR: teardown failed")
            if hold is None:
                hold = err

        try:
            self.log_event("About to run final post")
            self.final_post.run()
            self.log_event("final post completed")
        except Exception as err:
            self.log_event("ERROR: final post failed")
            if hold is None:
                hold = err

        if hold is not None:
            raise hold

    def run_test_set(self, test_set):
        hold = None
        try:
            self.log_event("About to run pre test", test_set)
            test_set["pre_test"].run()
            self.log_event("pre test completed", test_set)
        except Exception as err:
            self.log_event("ERROR: pre test failed", test_set)
            hold = err
        if hold is None:
            try:
                self.log_event("About to run test", test_set)
                test_set["test"].run()
                self.log_event("test completed", test_set)
            except Exception as err:
                self.log_event("ERROR: test failed", test_set)
                hold = err
            try:
                self.log_event("About to run post test", test_set)
                test_set["post_test"].run(test_outputs=test_set["test"].test_outputs,
                                          test_results=test_set["test"].test_results)
                self.log_event("post test completed", test_set)
            except Exception as err:
                self.log_event("ERROR: post test failed", test_set)
                if hold is None:
                    hold = err

        if hold is not None:
            raise hold

    def log_event(self, msg, test_set=None):
        now = datetime.now()
        time = now.strftime("%H:%M:%S")
        marker = "****"
        if test_set is not None and test_set["name"] is not None:
            msg = f"{msg} [{test_set['name']}]"
        print(marker)
        print(f"{marker} {time}: {msg}")
        print(marker)


# pylint: disable=too-many-arguments
class ClusterTestRunner(ClusterTestSetsRunner):
    """A simple cluster test runner that:
    . provisions a cluster
    . runs any pre_test (if provision was successful)
    . runs the test (if provisioned and any pre_test was successful)
    . runs post_test (if the test ran)
    . tears down the cluster"""

    def __init__(
        self,
        cluster=NullCluster(),
        pre_test=NullPreTest(),
        test=NullTest(),
        post_test=NullPostTest(),
        final_post=NullPostTest(),
    ):
        super().__init__(
            cluster=cluster,
            final_post=final_post,
            sets=[
                {
                    "name": None,
                    "pre_test": pre_test,
                    "test": test,
                    "post_test": post_test,
                }
            ],
        )

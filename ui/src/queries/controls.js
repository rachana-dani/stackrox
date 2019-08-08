import gql from 'graphql-tag';

export const AGGREGATED_RESULTS_ACROSS_ENTITY = gql`
    query getAggregatedResults(
        $groupBy: [ComplianceAggregation_Scope!]
        $unit: ComplianceAggregation_Scope!
        $where: String
    ) {
        results: aggregatedResults(groupBy: $groupBy, unit: $unit, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                numFailing
                numPassing
                unit
            }
        }
        controls: aggregatedResults(groupBy: $groupBy, unit: CONTROL, where: $where) {
            results {
                __typename
                aggregationKeys {
                    __typename
                    id
                    scope
                }
                numFailing
                numPassing
                unit
            }
        }
        complianceStandards: complianceStandards {
            id
            name
        }
    }
`;

export const AGGREGATED_RESULTS = gql`
    query getAggregatedResults(
        $groupBy: [ComplianceAggregation_Scope!]
        $unit: ComplianceAggregation_Scope!
        $where: String
    ) {
        results: aggregatedResults(groupBy: $groupBy, unit: $unit, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                numFailing
                numPassing
                unit
            }
        }
        controls: aggregatedResults(groupBy: $groupBy, unit: CONTROL, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                numFailing
                numPassing
                unit
            }
        }
        complianceStandards: complianceStandards {
            id
            name
        }
        clusters {
            id
            name
            namespaces {
                metadata {
                    id
                    name
                }
            }
            nodes {
                id
                name
            }
        }
        deployments {
            id
            name
        }
    }
`;

export const AGGREGATED_RESULTS_WITH_CONTROLS = gql`
    query getAggregatedResults(
        $groupBy: [ComplianceAggregation_Scope!]
        $unit: ComplianceAggregation_Scope!
        $where: String!
    ) {
        results: aggregatedResults(groupBy: $groupBy, unit: $unit, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                numFailing
                numPassing
                unit
            }
        }
        complianceStandards {
            id
            name
            controls {
                id
                name
                description
            }
        }
    }
`;

export const CONTROL_NAME = gql`
    query getControlName($id: ID!) {
        control: complianceControl(id: $id) {
            id
            name
            description
        }
    }
`;

export const CONTROL_QUERY = gql`
    query controlById($id: ID!, $groupBy: [ComplianceAggregation_Scope!], $where: String) {
        results: complianceControl(id: $id) {
            interpretationText
            description
            id
            name
            standardId
        }

        complianceStandards {
            id
            name
        }

        entities: aggregatedResults(groupBy: $groupBy, unit: CONTROL, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                keys {
                    ... on Node {
                        clusterName
                        id
                        name
                    }
                }
                numFailing
                numPassing
            }
        }
    }
`;

export const NODES_WITH_CONTROL = gql`
    query nodesWithControls($groupBy: [ComplianceAggregation_Scope!], $where: String) {
        entities: aggregatedResults(groupBy: $groupBy, unit: CONTROL, where: $where) {
            results {
                aggregationKeys {
                    id
                    scope
                }
                keys {
                    ... on Node {
                        clusterName
                        id
                        name
                    }
                }
                numFailing
                numPassing
            }
        }
    }
`;

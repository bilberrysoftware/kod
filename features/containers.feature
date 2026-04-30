Feature: Handle container lists and definitions
  In order to package everything required for a helm chart deployment
  As a Kubernetes Engineer
  I need to be able to handle container definitions and sets of definitions correctly

  Scenario: Remove duplicates from container lists
    Given there is a new Helm Chart Processing Result
    And a Helm Chart Processing Result container with Registry 'quay.io' and Repository 'kiwigrid/k8s-sidecar' and Tag '2.6.0' and Digest ''
    And a Helm Chart Processing Result container with Registry 'quay.io' and Repository 'wibble/flibble' and Tag '2.6.0' and Digest ''
    And a Helm Chart Processing Result container with Registry 'quay.io' and Repository 'kiwigrid/k8s-sidecar' and Tag '2.6.0' and Digest ''
    When I process the Helm Chart Processing Result to get a Container List
    Then there should be 2 containers in the container list
    And one of these containers with Registry 'quay.io' and Repository 'kiwigrid/k8s-sidecar' and Tag '2.6.0' and Digest ''
    And one of these containers with Registry 'quay.io' and Repository 'wibble/flibble' and Tag '2.6.0' and Digest ''
# Feature planning

Below are our future plans... Please 
[Log an Issue on GitHub](https://github.com/bilberrysoftware/kod/issues) 
if you have any strong opinions on our product direction.

Please read the [Contributions Guide](https://github.com/bilberrysoftware/kod/blob/main/CONTRIBUTING.md) 
and the
[Code of Conduct](https://github.com/bilberrysoftware/kod/blob/main/code-of-conduct.md)
first.

## Upcoming releases

### v0.2.0 release

- [X] Add in version CLI command support, and building of version into binary
  - Use GoReleaser - https://goreleaser.com/getting-started/intro/
  - [X] Change to use Semantic Versioning to be compatible with GoReleaser and other GoLang CLI norms
- [X] Bug container copy on package ignores other containers in same folder - E.g. docker.io - due to 'exists' check on folder instead of file
  - Was actually an incorrect error message that only showed the top level folder name - fixed to show full path to container tar file
- [X] Support helm -f / --values file flag (one or more values files) at deployment time
- [X] Invoke command on remote helm reference
  - [X] -r registry flag (this indicates a remote helm chart will be used)
  - [X] -c chart name flag
  - Note: uses helm repo add/update/save to accomplish this
  - [X] We may need to support the helm fetch flags for OCI repositories on the package command
    - Used cert-manager to test, but this package doesn't deploy due to CRDs not explicitly being added via values file
- [X] Change all exec commands to show their final output on the screen if it fails
- [X] Add a check for the helm command as this will be a prerequisite in this instance
- [X] Add deploy support for helm --wait
- [X] Ensure that the *nix TMPDIR=/some/other/folder options are being followed by os.TempDir() in GoLang - YES Obeyed
- [X] Skopeo invocation enhancements
  - [X] Stream output of skopeo copy so the user can monitor progress of the downloads
  - [X] Check for Skopeo on path before executing package command (Note that skopeo is not available on windows)
  - [X] Add --retry-times 5 flag to skopeo copy commands for resilience
  - [X] Check to see if skopeo needs to be logged in to access any of the containers mentioned (and error if so)
    - Only doing this for deploy, as we may well have anonymous access to fetch container images on package
  - [X] Fetch container info to determine digest, and save this to the container info in the package file
    - via `skopeo inspect docker://someregistry/myimagename` or oci:// URL
    - Add in SkopeoInspectResult YAML as a result type
    - [X] Works, but cannot load images without imagePullSecrets being supported, so need to do that now
      - Prefer global.imagePullSecrets over imagePullSecrets
      - Added anyway: No-one is dumb enough to default this in a chart, surely?: Ensure we add to this, rather than replace it
      - Mostly working, but still some references we're not picking up on:-
        - quay.io/prometheus/node-exporter:v1.10.2
        - registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.18.0
        - quay.io/kiwigrid/k8s-sidecar:2.5.0
        - docker.io/grafana/grafana:12.4.2
        - Find out what elements these live in, and if we can reasonably detect them without a hints file
          - They're all sub chart dependencies!
  - [ ] Further image detection improvements
    - [ ] Another convention is global.imageRegistry (2 of top ten) (Will override other registry refs)
    - [X] Search entire tree for '.*\[iI\]mage' and sub elements (4 of top ten)
      - ArgoCD, kube-prometheus-stack
    - [ ] Allow specification of container images, and links to config elements, for things we cannot guess - kod-hints.yaml
      - Note: include this as kod-hints.yaml in the output archive too
    - [ ] Add flat elements without image parent. E.g. imageRegistry and imageRepository and imageNamespace
      - cert-manager
    - [ ] Consider reading images required from the deployment YAML templates (helm template), and reverse parse their parameters (Complex)
      - If it works, we should be able to support any helm chart OOTB without a hints file
  - [ ] Add subchart support
    - [ ] Separate chart packaging into its own function 
    - [ ] Read dependency chart information
      - Required for all Bitnami charts that use the bitnami-common dependency (Redis, MongoDB are known)
      - Also needed for kube-prometheus-stack (Grafana et al)
      - We may need to rewrite the Chart.yaml here, or figure out how to override the source chart URLs to OCI URLs
    - [ ] Call chart packaging function recursively
    - [ ] Ensure we check to see if the same chart has already been downloaded
    - [ ] Check through deploy and helm invocation to see how it finds the helm chart reference locally if cached
  - [ ] Add archiving and unarchiving progress reports using `archives` module (Otherwise it feels like the CLI has hung)
  - [ ] Improve container version detection by checking `inspect` output RepoTags list for the version, or latest, or first listed (most recent) version before executing `skopeo copy`
- [ ] Allow packaging with -f option to override default values, especially image paths to a local container registry
  - Note that this is useful where you're already copying container images from public repos to private repos, as you'll want to package those versions up, and will likely already have the values override file to specify the right container images 
  - [ ] Use this with helm template to parse final output to find all container image references
    - How to reverse engineer this back into properties that can be overridden?
  - [ ] Include these as CHARTNAME-CHARTVERSION-001-values.yaml and 002, 003 etc within the package
    - This is particularly useful for providing default, well known, configuration to data layer helm charts (Kafka, Nifi, elasticsearch, MongoDB et al)
    - Kept as separate files to aid in debugging and security verification of kod
  - [ ] Update deploy to use these values files
- [ ] Helm enhancements
  - [ ] Check to see if helm needs to be logged in when fetching packages, or errors due to this on package
  - [ ] Consider adding -o option on deploy to specify where in the container/artifact registry to save helm artifacts to
- [ ] Provide a --hide-registry-urls option on package so that an internal replica of public container images doesn't end up named 'myinternalregistry.local/docker.io/MYNAME:TAG' and is instead just named for its original location 'docker.io/MYNAME:TAG'
  - Note that the image itself may still include this reference as a separate tag value
- [ ] Consider splitting deploy into populate and deploy commands so we have a separation of concerns between those who can write to the repo and deploy the package
  - Still make just executing deploy check if it needs to populate, if executed from a .kodpkg
  - Document the different personae of users / actors in invoking kod
  - Have populate command create a mypackage-myversion.kodapp file that describes where locally content resides, and that can be passed to a deployer and used in the deploy command
    - By definition, this should include a helm values file with the necessary image elements overridden
    - Should be another archive format, with an updated kod-package.yaml and a new kod-app.yaml and kod-values.yaml included
      - We may not need a kod-app.yaml file, but instead update the container image references in kod-package.yaml, and change helm reference to an OCI artifact URL
- [ ] Add better error handling to invocation of helm upgrade --install as it currently hangs even without --wait, E.g. if one pod doesn't come up right
- [ ] Add examples to the help output of all kod commands. E.g. chart locally, chart on OCI url etc.
- [ ] Apply Linux Foundation standards to repo - code of conduct, contributing, security reporting, etc.
- [ ] How do we know we have succeeded?
  - [ ] Bitnami Redis chart works (Chart Dependencies)
  - [ ] Bitnami MongoDB chart works (Chart Dependencies)
  - [ ] Cloudnative Postgres chart (NOT Bitnami version) - https://github.com/cloudnative-pg/charts
  - [ ] Cert-manager chart works (Remote chart, non-standard image location)
    - imageRegistry, imageNamespace and image.name, weirdly, OR image.repository to override ALL THREE
  - [ ] Nifikop operator works (Remote chart + OCI) `go run main.go deploy -n nifi-operator -p /tmp/kod-nifikop-1.16.0.kodpkg -r  https://REGISTRY:8080/ -d nifikop -f examples/charts/nifikop/basic-values.yaml`
  - [ ] kube-prometheus-stack chart works (*mage image refs, and chart dependencies, and sidecar containers in those dependencies)
- [ ] Update docs with helm charts that work
  - [ ] Don't forget to include notable subcharts in that list, and perhaps invoke those that make sense on their own (E.g. Grafana)
- Note: BREAKING CHANGES
  - Changed Hints file hints element to images, as they're only container image hints. Added imagePullSecrets as another type of hint

### v0.3.0  release - usage checks and environment support

- [ ] Ensure Istio works with kod (Because I personally really want to use it with Istio!)
- [ ] Add in checks for all required/optional values for each command, and sensible values checks
- [ ] Enforce best practice. E.g. no 'latest' tag versions, and add --allow-latest-tag flag
- [ ] Error if container image being copied matches an existing one but has a different Digest value
- [ ] Check for different versions of the same container image being used, and output a warning to CLI
- [ ] Allow specification of target architecture, or all in the same - default to amd64 and allow --arch=all and --arch=arm64 etc?
  - If specified, place name of target architecture in the archive file name
  - Add specification of container architecture for deployment, if supported by skopeo (i.e. from 'all' archs archive)

### v0.4.0 release - more complex chart installation use cases

- [ ] Add single helmfile support
  - [ ] Check for helmfile CLI on the path
  - [ ] Basic single helm chart support
  - [ ] Multiple helm charts, multiple namespaces support
  - [ ] Same helm chart, multiple namespaces support (E.g. Istio ingress/egress gateways)
  - [ ] helmfile overrides file support (deploy -f option may need its docs changing)
- [ ] Add Chart dependency chart support
  - Will require a deploy:list element in the kod-package file to specify the installable ones that are NOT just dependencies
    - This has the side effect of potentially opening up the possibility to include and invoke multiple helm charts/helmfiles
    - This must be avoided, otherwise semantics like -f to specify override helm chart values files and -n for namespace make no sense
    - Prefer multiple invocations of kod and separate packages for distinct charts

### v0.5.0 release - Usability and compliance improvements

- [ ] Add `package-tools` support - To enable all offline external components to be packaged
  - Note: Linux amd64 support only initially
  - [ ] Include skopeo, helm3, kubectl, and of course kod itself
  - [ ] Include a helper install.sh in the package
  - [ ] Add `deploy-tools` support that unpacks and executes install.sh
    - Note: Would work with any tgz with an install.sh file in it
      - Add auto archive format detection support to deploy command to support this use case
  - [ ] Add `unpack-tools` support to unpack but not install the archive
  - [ ] Add --local flag to copy ones on the local system instead of downloading the latest from the interwebs
    - Or do the opposite by adding a --latest flag to download the latest only
  - [ ] Consider adding --extras flag to include some of my favourite tools - kube-capacity, zotcli
  - [ ] Add MacOS support (arm64)
- [ ] Add package-zot to package up zot binaries and cli and an install.sh script - Linux Ubuntu amd64 only
- [ ] Add bash completion support - See https://cobra.dev/docs/how-to-guides/shell-completion/
- [ ] Add --ci flag to not prompt with questions (E.g. skopeo login)
- [ ] Add --output-commands-only to deploy to only output the manual bash commands to carry out the necessary actions, and no other logged information, without carrying the commands out.
- [ ] Generate SBOM files during package generation
- [ ] Output sha256 digest file alongside archive
- [ ] Log sha256 hashes of components within a sumfile in the archive itself too, and check this on all command invocations
  - Can also be used to see if unpacking commands need re-running on partial unpack/deploy failures
  - Fail on deploy if the sha doesn't match
- [ ] Add sha256sum support to package-tools, deploy-tools, package-zot, deploy-zot too

### v0.6.0 release - Ease of use enhancements, and OCI support

- [ ] Add common mappings for the top ten charts' image specs into the kod binary
- [ ] Add colour terminal and emoji support to the output in the terminal
- [ ] Add --cleanup flag to package and deploy to clean up temporary folder after all operations (except unpack)
- [ ] Add --sanitise flag to remove the internal container registry names from the helm charts and hints files, and generate our own
- [ ] Add support for local .tgz helm chart archives for `package -c`
- [X] Done on Mon 06 Apr 2026 - Add support for `oci://` URLs for helm chart archives via `package -c`
- [ ] OCI artifact support
  - See https://zotregistry.dev/v2.0.0/articles/workflow/ for an example OCI artifact workflow 
  - [ ] Add a populate command which is basically deploy with --no-install specified 
  - [ ] Store deployed/populated helm charts in a target OCI repository, if the registry supports OCI artifacts
  - [ ] Store a kodpkg in the target OCI registry directly (I.e. `populate --no-unpack`)
  - [ ] Rework the kodpkg format as required such that the hints files and values overrides are within their own artifact
    - This may well actually be the kodinfo package format for a helm chart
  - [ ] Ensure `deploy -c` supports an OCI URL to a kodinfo OCI artifact archive
    - Note: Maintain the old file-system based functionality so that we continue to support non-OCI registries or target environments
- [ ] Add support for artifactory URLs for helm chart archives via `--rt` flag and the `jf rt dl` command, if jf is installed
  - Do this as an addon?
  - Also need to update the package-tools command to include --rt and the jfrog cli

### v0.7.0 release - Release automation

- [ ] Create GitHub Actions build and CI linting check support
- [ ] Create Concourse public builds and testing pipeline
  - Include support for Windows, arm64, raspberry pi arm architectures
- [X] Done on Mon 06 Apr 2026 - via GoReleaser and GitHub Actions - Add support for GoLang builds onto multiple environments
- [X] Done on Mon 06 Apr 2026 - via GitHub Actions instead - Push out a v1.0-rc1 release package using Concourse as a pipe cleaning exercise
  - Note: It's a manual step to tag a release, so just run Concourse tests after a commit to main but before a tagging to ensure all platforms pass before release

### v0.8.0 release - Ensure V1 is ready for prime time

- [ ] Add in community required files, security procedure, etc
- [ ] Ensure new Bilberry website is live, with a page about kod, and links to/from kod repo
- [ ] Ensure build badges are included in the README.md file
- [ ] Add automated creation of an Issue (bug) if a CD build or integration test fails
- Anything glaring I've missed...

### v1.0.0 full release

- Should hopefully just be a more solid build of rc2, with automated publishing

### v1.1.0 release - Helm chart and helmfile consumption and kod package generation aids

- [ ] Self-publishing support - Consider adding a way to allow third party helm/helmfile developers to include default values and hints file without us creating it
  - E.g. a .kodinfo package containing a kod-package.yaml, one or more MYCHART-MYCHARTVER-kod-hints.yaml files
  - Will require an update to the package command E.g. -i <mypackage.kodinfo>
  - Should support fetching over public https (E.g. to published website or GitHub direct file URL)
  - Consider how this would compose if, for example, someone created two helm charts with a kodinfo and one helmfile referencing them with its own kodinfo

### Descoped

- [ ] Refresh command with new version of helm charts
- [ ] Helper to install skopeo if it doesn't already exist when executing the package command (sudo apt-get update; sudo apt-get -y install skopeo)
- [ ] Consider adding support for curl/wget instead of `jf rt dl` where that CLI is not available for `kod package --rt`

## Previous releases

## v0.0.1 (Aka MVP v1.0-alpha1) - Provides limited value to deploy some basic helm charts without dependencies

- [X] Invoke kod command pointing at a Helm chart (local folder for now)
- [X] Write a summary description document in a temp folder (lock file)
- [X] Package temp folder as a tar.xz archive (.kodpkg extension) (And re-use a previous temp folder)
- [X] Read a helm chart and determine the direct images required by examining the deployment variables
    - Note: Using the Bitnami chart repo as a guide to common approaches. See https://github.com/bitnami/charts/tree/main/bitnami
    - [X] Single image is easy - convention on helm create is image {repository, tag}
    - [X] Next is image {registry, repository, tag, digest}
    - [X] Support Chart.appVersion when tag isn't specified
- [X] Invoke Skopeo to fetch and package the container images, keeping their original SHA hash
    - See https://github.com/containers/skopeo/blob/main/README.md
    - [X] Add support to default back to 'latest' if it fails
- [X] Test against a variety of known helm charts - 5 working so far out of 13 tested
    - [X] Kafka strimzi operator (OCI only - cannot test yet)
        - helm install my-strimzi-cluster-operator oci://quay.io/strimzi-helm/strimzi-kafka-operator
    - [X] Bitnami Redis - WORKS
    - [X] Redis operator
        - helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
        - helm install redis-operator ot-helm/redis-operator --namespace ot-operators --set featureGates.GenerateConfigInInitContainer=true
        - redisOperator.imageName (includes registry), imageTag (blank, defaults to Chart.AppVersion)
    - [X] Nifi Kop operator - WORKS
        - https://konpyutaika.github.io/nifikop/docs/2_deploy_nifikop/2_customizable_install_with_helm
        - standard image repository and tag
    - [X] cloudnative-pg operator - WORKS
    - [X] elastic kubernetes operator
        - Strange one: image: "{{ .Values.image.repository }}{{- if .Values.config.ubiOnly -}}-ubi{{- end -}}{{- if .Values.image.fips -}}-fips{{- end -}}:{{ default .Chart.AppVersion .Values.image.tag }}"
        - Defaults tag to Chart.appVersion - was not defaulted because tag was 'null' string - FIXED
        - Also has image.fips=false by default
    - [X] Bitnami MongoDB - WORKS
    - [X] Most popular on artifact hub. See https://artifacthub.io/stats on right hand side
        - [X] kube-prometheus-stack
            - oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
            - https://prometheus-community.github.io/helm-charts with chart kube-prometheus-stack
            - https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack
            - uses global.imageRegistry
            - ?.alertManagerSpec.image {registry, repository, tag, sha}
            - ?.(prometheusOperator).image {registry, repository, tag, sha}
            - ?.patch.image {registry, repository, tag, sha}
        - [X] argo-cd
            - https://github.com/argoproj/argo-helm/tree/main/charts/argo-cd
            - Uses global.image {repository (includes registry), tag}
            - Also controller.image, dex.image, dex.initImage, redis.image, redis.exporter.image, redis-ha as left too
            - redisSecretInit.image (both empty by default)
            - server.image, server.extensions.image
        - [X] cert-manager
            - https://charts.jetstack.io/cert-manager
            - Uses imageRegistry and imageNamespace
            - Also deprecated image.name image.repository and image.tag (which is commented out and non discoverable)
            - tag defaults to Chart.appVersion
        - [X] postgresql (This is Bitnami Postgresql) - WORKS
        - [X] redis (This is Bitnami Redis)
        - [X] loki
            - https://github.com/grafana-community/helm-charts/tree/main/charts/loki
            - uses global.imageRegistry
            - {loki,lokiTest,lokiCanary,gateway,write,read,backend,ingestor,distributor,querier}.image {registry, repository, tag, digest}
            - gateway.metrics.image {repository (includes registry), tag}
            - [ ] tag and digest are 'null' and not an empty string - need to validate this
                - sometimes registry and repository are null too ((write,read,backend,ingestor,distributor,querier).image)
            - PROBABLY MORE, I GAVE UP!
        - [X] keycloak (This is Bitnami Keycloak)
            - NO TAGS - Looks abandoned or moved from docker.io
        - [X] prometheus
            - OCI Artifact: oci://ghcr.io/prometheus-community/charts/prometheus
            - Helm Repository: https://prometheus-community.github.io/helm-charts with chart prometheus
            - https://artifacthub.io/packages/helm/prometheus-community/prometheus
            - server.image {repository (includes registry), tag (blank, defaults to Chart.appVersion), digest (blank)
                - Note: Also checks for server.image.distroless and appends '-distroless' if it's true
- [X] Deploy a package (unpack to temp, invoke helm upgrade --install)
    - Note: Do an installation via helm template so we can verify we can replace all the image: values used in container spec
    - [X] Add unpack command to create temp folder first
    - [X] Process archive and list containers
    - [X] Execute skopeo command to upload container image
        - Turns out Skopeo requires zot to use https only for authentication and integration
    - [X] zot testing: Set up TLS cert and add to Microk8s trusted cert list so installation works, and skopeo verifies TLS
- [X] Support -r flag for the local registry base to use during deployment
    - [X] Test against an Ubuntu Service installation of zot
        - [X] Install and verify that zot is running and accessible
        - [SKIP] Add --insecure-no-verify command option for bypassing TLS server checks
            - Note that skopeo does this the other way around, so we should add --verify by default to skopeo
            - Skopeo must run against a valid https server with a valid cert - no way around it
- [X] Validate helm and kod installations are identical with the same values file (except image URL obvs)
    - [X] Bitnami Redis
        - Fails on requiring dependency in Chart.yaml
        - Bitnami Common: registry=oci://registry-1.docker.io/bitnamicharts, tag=bitnami-common
    - [X] Bitnami Postgresql
        - Same issue as above
    - [X] Bitnami MongoDB
        - Same issue as above
    - [X] Nifi Kop operator
        - Requires cert-manager
            - Error: unable to build kubernetes objects from release manifest: [resource mapping not found for name: "nifikop-webhook-cert" namespace: "default" from "": no matches for kind "Certificate" in version "cert-manager.io/v1"
            - ensure CRDs are installed first, resource mapping not found for name: "selfsigned-issuer" namespace: "default" from "": no matches for kind "Issuer" in version "cert-manager.io/v1"
    - [X] cloudnative-pg operator
        - Worked first time! Sun 06 Apr 2026 07:07
        - Still using remote container image reference - need to rewrite that in the helm chart to the local value
    - [X] Istio charts (Because I know these are standalone without dependencies)
        - [X] Download source of helm chart first - we cannot use remote chart references yet
        - [X] istio-base
            - Worked first time! Sun 06 Apr 2026 07:20
            - Still using remote container image reference - need to rewrite that in the helm chart to the local value
        - [X] istiod
            - Worked first time! Sun 06 Apr 2026 07:20
            - Still using remote container image reference - need to rewrite that in the helm chart to the local value
        - [X] istio ingress gateway
            - Worked first time! Sun 06 Apr 2026 07:23
            - Still using remote container image reference - need to rewrite that in the helm chart to the local value
        - Note: Istio won't actually find the correct container reference, as istiod dynamically injects it at runtime - We'll need to figure this out
        - Note: Egress gateway needs --set or -f support
    - [X] Turn these tests into scripts to be executed by a CD env (but that can be ran manually)
    - Might be required to get some charts working at all
- [X] Support local container images by pointing references in YAML to new internal registry
    - [X] Package: Output hints file within charts folder next to its chart CHARTNAME-CHARTVER-hints.yaml
    - [X] Deploy: Generate values file with new deployment image names and location using hints file, to CHARTNAME-CHARTVER-values.yaml in chart folder
- [X] Basics for v1.0-alpha1 release
    - [X] Add basic documentation
    - [X] Ensure docs are built and pushed to the oss.bilberrysoftware.com/kod subsite
        - Created as 'kod-website' artifact in GitHub Actions workflow
    - [X] Perform manual build and test of kod CLI ready to create Release bundle
- [X] Complete and push this version

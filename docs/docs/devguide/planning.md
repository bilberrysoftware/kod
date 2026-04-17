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
  - [X] Further image detection improvements
    - [X] Add support for cert-manager imageRegistry and imageNamespace and COMPONENT/image.name
      - Has old and new way. New way is imageRegistry/imageNamespace/{component}:image.name:{image.digest (NOT DEFINED IN VALUES.YAML) OR image.tag (NOT DEFINED IN VALUES.YAML) OR Chart.AppVersion}
      - [X] Package working
      - [X] Deploy working (Looks for ^ in hint to separate parts of Repository out among two global fields)
        - [X] Ensure sub-element search does process multiple element values (i.e. firstlevel.secondlevel.valuefield)
    - [X] Search entire tree for '.*\[iI\]mage' and sub elements (4 of top ten)
      - ArgoCD, kube-prometheus-stack
  - [X] Add subchart support
    - [X] Separate chart packaging into its own function 
    - [X] Read dependency chart information
      - Required for all Bitnami charts that use the bitnami-common dependency (Redis, MongoDB are known)
      - Also needed for kube-prometheus-stack (Grafana et al)
      - Read the Chart.lock file, if present, for precise version info
      - [X] Sim link MYKOD/charts/MYCHART/charts to MYKOD/charts so we don't double-download any helm chart
        - It looks like helm pull is already creating these sub folders when pulled from an OCI repo
    - [X] Call chart packaging function recursively
    - [X] Verify all charts and containers are downloaded, if it's possible to detect them
      - [X] BUG package command is not fully populating any found containers from sub charts
    - [X] Ensure we check to see if the same chart has already been downloaded
      - [X] BUG: If oci helm chart unpack temporary folder exists, instead of not downloading and just processing the folder, it tries to redownload and fails
        - Still errors:-
          - go run main.go package -r oci://ghcr.io/konpyutaika/helm-charts -c nifikop
            Using chart registry: oci://ghcr.io/konpyutaika/helm-charts
            WARNING: You are using an OCI Registry (-r) URL with a chart path (-c). Normally, you just use the full OCI URL with the -r option. We've rewritten your OCI URL to: oci://ghcr.io/konpyutaika/helm-charts/nifikop
            Executing /usr/sbin/helm pull oci://ghcr.io/konpyutaika/helm-charts/nifikop --untar --untardir /tmp/kod-package-c182428c8065d47728e8a5736dc5d5b5ddd207d660d2bc194bd6a81a5154ef58
            Error executing OCI helm pull: exit status 1 details: []
            exit status 1
      - [X] BUG: Package of cert-manager fails - double download folder cert-manager/cert-manager
        - go run main.go package -r https://charts.jetstack.io -c cert-manager -f chartfiles/cert-manager/values.yaml
        - Was prefixing cert-manager twice. Moved to same logic as OCI repos
      - [X] Need to verify nifikop deploy once cert-manager is installed (requires Issuer CRD)
        - go run main.go package -r oci://ghcr.io/konpyutaika/helm-charts -c nifikop (Works first time)
        - go run main.go deploy -p /tmp/kod-nifikop-1.17.0.kodpkg -s zot -d nifi -r $REGISTRY
        - [X] Needs image.imagePullSecrets set, and this wasn't detected on package creation
        - [X] BUG: Figure out why it's not running - may needs its own default values file - Just missing its own runtime config
      - [X] Ensure subsequent runs of package without removing temporary files always runs, and doesn't break hints files or other metadata files
        - Still errors:
          -  go run main.go package -r oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
             Using chart registry: oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
             Executing /usr/sbin/helm pull oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack --untar --untardir /tmp/kod-package-303c70341f8b052810911aa419b3459dc4b447043816ead6e27470374fffb9ae
             Error executing OCI helm pull: exit status 1 details: []
             exit status 1
        - Due to file already existing:-
          -  /usr/sbin/helm pull oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack --untar --untardir /tmp/kod-package-303c70341f8b052810911aa419b3459dc4b447043816ead6e27470374fffb9ae
             Pulled: ghcr.io/prometheus-community/charts/kube-prometheus-stack:83.4.0
             Digest: sha256:073a9207738d71cd36911847711b2a2b9aecc9e1e979b8e9bb164ca937ae6cf9
             Error: failed to untar: a file or directory with the name /tmp/kod-package-303c70341f8b052810911aa419b3459dc4b447043816ead6e27470374fffb9ae/kube-prometheus-stack already exists
    - [X] BUG: Container list creation for unique containers still produces duplicates
      - [X] Write Cucumber test harness for this
    - [X] BUG: When the skopeo inspect corrects the image tag, this isn't reflected in the saved image tar name, causing the container to not be found on installation
    - [X] BUG: Only generating hints file for main chart and windows exporter during package
    - [X] Image location override values file for main chart does not yet override subchart values during deploy
      - [X] Check that duplicate values in root hints file (grafana.image) isn't an issue on deploy
        - [X] Seems to be fine, but log a bug to fix this at some point
      - [X] Yes: Ensure deploy handles hints files with parents with long multi-element paths (should do...)
      - [X] BUG: packagedImage tag value appears to be being used in values instead of actual downloaded tag - likely due to container fetch fallback rules not being reflected in HintsFile (correctly)
      - [X] BUG: Something in kube-prometheus-stack is causing hint to not be found for:-
        - image: registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.18.0
        - image: quay.io/prometheus/node-exporter:v1.11.1
        - Note: This now works, but the default values file results in a broken installation! See https://oneuptime.com/blog/post/2026-02-09-kube-prometheus-stack-grafana/view
        - [X] BUG: container map always has source data, but hints now has target data, and so versions are not found
      - [X] BUG: When version and digest are both present (all the time now that package populates it), the URL is invalid for fetching
        - I think you need either the digest or the tag, not both
        - We need to ideally default to digest, assuming they are the correct digest values
        - Change made, but didn't affect kube-prometheus-stack prometheus-node-exporter pod as that chart does not handle both tag and digest being present
          - We'll have to create a hints file for kube-prometheus-stack that removes the prometheus-node-exporter.image.tag value (set to "")
    - [X] BUG: Missing container image on disc causes deploy to always fail - should WARN and refer to compliance warnings file
    - [X] Change temp folder to kod-deploy for deploy and kod-package for package so they don't overlap when testing on the same machine
      - [X] Double check that kod package doesn't put '-package-' in the final package filename
    - [X] Check through deploy and helm invocation to see how it finds the helm chart reference locally if cached
      - Uses a lockfile - we need to write that
      - Update lockfile (rename to Chart.lock.orig), and generate our own with the new OCI chart URLs
      - No need as we're using symlinks to link to the correct version
  - [X] Add archiving and unarchiving progress reports using `archives` module (Otherwise it feels like the CLI has hung)
    - [X] Added mechanism to Unarchive, need to test (based on fi.Size() - which may be raw or archived size... TBD)
      - [X] Verify that Unarchive always printing out its progress isn't confusing. (Also used for unpacking chart tgzs during package)
      - Unpacking takes a while to start before progress of 0% is shown
      - Progress jumps from 0% to 100% with no intermediate progress points
      - ABANDONED - absolutely eats CPU cycles. This is referenced in tar file handling in the repo. May be better to use tar xf when simply unpacking
        - [X] DONE Consider using raw `tar xf FILENAME` to unpack... See if we can get a progress output too.
    - [X] Add equivalent to Archiving of the final kodpkg file
- [X] BUG: Output of long commands with \n in them such as helm upgrade --install doesn't prepend emoji on every line
- [X] Allow packaging with -f option to override default values, especially image paths to a local container registry
  - Note that this is useful where you're already copying container images from public repos to private repos, as you'll want to package those versions up, and will likely already have the values override file to specify the right container images 
  - [X] NOT NEEDED YET - Consider using this with helm template to parse final output to find all container image references
    - How to reverse engineer this back into properties that can be overridden?
  - [X] Include these as CHARTNAME-CHARTVERSION-001-values.yaml and 002, 003 etc within the package
    - This is particularly useful for providing default, well known, configuration to data layer helm charts (Kafka, Nifi, elasticsearch, MongoDB et al)
    - Kept as separate files to aid in debugging and security verification of kod
  - [X] Update deploy to use these values files
  - [X] Test with cert-manager as it needs crds.enabled=true for the startupapicheck to work on helm install - make this our first built-in
- [X] Add deploy --create-secret to populate the imagePullSecret in the relevant namespace (if skopeo logged in, create from the ~/.docker/config.json)
  - [X] create this in parallel with the helm deploy, so that the user doesn't have to create the namespace manually before helm install, and so you don't have to wait for helm install --wait (as containers will fail until the NS exists)
- [X] Helm enhancements
  - [X] Add package -c support for Helm TGZ URLs without a repo. E.g. https://github.com/cloudnative-pg/charts/releases/download/cloudnative-pg-v0.28.0/cloudnative-pg-0.28.0.tgz
- [ ] Add more top ten chart support
  - [X] Standalone postgres instance
    - [X] Default docker host to docker.io when its blank (required by postgres chart)
      - [X] docker.io missing from container uploaded to container registry, and in deployment URL
        - values file tagpath is blank - image.""
        - note: versionOverride instead of tag/version (Is ok because it defaults to Chart.AppVersion)
      - [X] imagePullSecrets at top level not detected - if empty, is an []interface{} not a []string
  - [X] Retest kube-prometheus-stack - Fails as node-exporter uses both tag and digest if provided, and our digests don't match the original source yet
    - go run main.go package -r oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack
    - go run main.go deploy -p $TMPDIR/kod-kube-prometheus-stack-83.6.0.kodpkg -n monitoring -d kps -s zot --create-secret --no-digests -r $REGISTRY 
    - [X] BUG: Package doesn't pick up global.imagePullSecrets: [] anymore and doesn't inject it into the hints file
    - BUG: Doesn't like version number and SHA in ref - but only because the sha doesn't match the one in the repo
      - Back-off pulling image "$REGISTRY/kod/containers/quay.io/prometheus/node-exporter:v1.11.1@sha256:0f422f62c15f154af8d8572b23d623aebfb10cec73a5c654d18f911f3f9df241"
      - Likely due to this:-
        - Processing container with repository prometheus/node-exporter and tag v1.11.1
          WARNING Hint has an invalid container tag. Setting to this available container tag. Reg: quay.io Repo: prometheus/node-exporter Tag: 1.11.1 -> v1.11.1
          WARNING Hint has an invalid container tag. Setting to this available container tag. Reg: quay.io Repo: prometheus/node-exporter Tag: 1.11.1 -> v1.11.1
          Executing /usr/bin/skopeo copy docker-archive:/tmp/kod-kube-prometheus-stack-83.6.0/containers/quay.io/prometheus/node-exporter/v1.11.1.tar docker://$REGISTRY/kod/containers/quay.io/prometheus/node-exporter:v1.11.1 --retry-times 5
        - [ ] LATER: Packaging: kube-prometheus-stack: Requires the sha256 consistency fix before we can fix this
          - Now with --preserve-digests we get this:-
            - Executing /usr/bin/skopeo inspect docker://docker.io/library/busybox:1.37.0
              Executing /usr/bin/skopeo copy docker://docker.io/library/busybox@sha256:1487d0af5f52b4ba31c7e465126ee2123fe3f2305d638e7827681e7cf6c83d5e docker-archive:/tmp/kod-package-kube-prometheus-stack-83.6.0/containers/docker.io/library/busybox/1.37.0.tar --retry-times 5 --preserve-digests
              🚢 Getting image source signatures
              🚢 Copying blob sha256:481282afbc4304ffee4792258ea114f09e423a4a082335b30695b50310394f47
              🚢 Copying config sha256:925ff61909aebae4bcc9bc04bb96a8bd15cd2271f13159fe95ce4338824531dd
              🚢 Writing manifest to image destination
              FATA[0003] copying system image from manifest list: writing manifest: Unsupported manifest type, need a Docker schema 2 manifest
              Error running skopeo copy: exit status 1 details: Getting image source signatures
              Copying blob sha256:481282afbc4304ffee4792258ea114f09e423a4a082335b30695b50310394f47
              Copying config sha256:925ff61909aebae4bcc9bc04bb96a8bd15cd2271f13159fe95ce4338824531dd
              Writing manifest to image destination
              Trying backup approach of using tag 'latest' (Needed for many Bitnami container images)
              Executing /usr/bin/skopeo copy docker://docker.io/library/busybox:latest docker-archive:/tmp/kod-package-kube-prometheus-stack-83.6.0/containers/docker.io/library/busybox/1.37.0.tar --retry-times 5 --preserve-digests
              🚢 Getting image source signatures
              🚢 Copying blob sha256:481282afbc4304ffee4792258ea114f09e423a4a082335b30695b50310394f47
              🚢 Copying config sha256:925ff61909aebae4bcc9bc04bb96a8bd15cd2271f13159fe95ce4338824531dd
              🚢 Writing manifest to image destination
              FATA[0003] copying system image from manifest list: writing manifest: Unsupported manifest type, need a Docker schema 2 manifest
              Error running skopeo copy using latest tag: exit status 1 details: Getting image source signatures
          - [X] Workaround: Add a --nodigests flag to deploy for now, and don't do --preserve-digests?
          - Works but node-exporter is complaining (likely due to prometheus missing?):-
            - Error: failed to generate container "1a0fe2a06fccbcd901655a4c0f1e8c4bb047a7ed819181a59319604b34dede66" spec: failed to generate spec: path "/" is mounted on "/" but it is not a shared or slave mount
  - [X] Retest argo-cd
    - go run main.go package -r oci://ghcr.io/argoproj/argo-helm/argo-cd
    - go run main.go deploy -p $TMPDIR/kod-argo-cd-9.5.2.kodpkg -n cd -d myargo -s zot --create-secret -r $REGISTRY
    - [X] NIGGLE: If v1.2.3 not found, remove v rather than add another in tag for version guessing: docker://ecr-public.aws.com/docker/library/haproxy:vv3.3.7
    - [X] BUG: Package doesn't pick up global.imagePullSecrets: [] anymore and doesn't inject it into the hints file
  - [X] Retest traefik - Fails on imagePullSecrets detection
    - go run main.go package -r oci://ghcr.io/traefik/helm/traefik
    - go run main.go deploy -p $TMPDIR/kod-traefik-39.0.8.kodpkg -n traefik -d myt -s zot --create-secret -r $REGISTRY
    - [ ] LATER: Packaging: Traefik: Requires deployment.imagePullSecrets support
  - [X] Retest redis OT Operator - Fails on both image and imagePullSecrets detection
    - go run main.go package -r https://ot-container-kit.github.io/helm-charts/ -c redis-operator
    - go run main.go deploy -p $TMPDIR/kod-redis-operator-0.24.0.kodpkg -n redis-operator -d myro -s zot --create-secret -r $REGISTRY
    - [ ] LATER: Packaging: OT Redis Operator: Requires redisOperator.{imageName,imageTag (default: Chart.AppVersion),initContainerImageTag,imagePullSecrets:[]}
  - [X] Retest loki - Fails as the Chart's values.yaml doesn't produce a complete default configuration
    - go run main.go package -r oci://ghcr.io/grafana-community/helm-charts/loki
    - go run main.go deploy -p $TMPDIR/kod-loki-13.1.2.kodpkg -n loki -d myl -s zot --create-secret -r $REGISTRY
    - [ ] LATER: Packaging: loki: NIGGLE: Uses null instead of "" for global.imageRegistry (but each container has their own anyway)
    - [X] BUG: imagePullSecrets at top level was not detected
      - Because it has a global but this is NOT where imagePullSecrets sits! 
    - [ ] LATER: Packaging: loki: Create values file for default basic installation
      - Error: execution error at (loki/templates/monolithic/statefulset.yaml:48:8): Please define loki.storage.bucketNames.chunks
  - [ ] Retest keycloak (if non-Bitnami and non-OLM can be found)
  - [ ] Retest prometheus standalone
    - go run main.go package -r oci://ghcr.io/prometheus-community/charts/prometheus
    - go run main.go deploy -p $TMPDIR/kod-prometheus-29.2.1.kodpkg -n prometheus -d myprom -s zot --create-secret --no-digests -r $REGISTRY
      - [?] Deploy doesn't work due to below issue, but also that 3 containers have tags and digests, and digests don't match yet
    - [?] BUG: Packaging has two registry URLs for one container:-
      - Executing /usr/bin/skopeo inspect docker://docker.io/quay.io/prometheus/pushgateway:v1.11.2
        FATA[0001] Error parsing image name "docker://docker.io/quay.io/prometheus/pushgateway:v1.11.2": reading manifest v1.11.2 in docker.io/quay.io/prometheus/pushgateway: requested access to the resource is denied
        Error inspecting skopeo image: /tmp/kod-package-prometheus-29.2.1/containers/docker.io/quay.io/prometheus/pushgateway/v1.11.2.tar at url: docker://docker.io/quay.io/prometheus/pushgateway:v1.11.2 error: exit status 1 details:
      - Likely because it's registry is blank and we didn't check if registry path was blank instead!
        - WARNING: no version tag found for container. Defaulting to 'Chart.appVersion' for quay.io/prometheus/pushgateway
          Found container image: '/quay.io/prometheus/pushgateway:v1.11.2@' at path 'image'
          Searching for any element named '.*[iI]mage'...
          WARNING: no version tag found for container. Defaulting to 'Chart.appVersion' for quay.io/prometheus/pushgateway
      - [?] Added a more sophisticated domain name check at start of ctr.Repository if ctr.Registry == "" to prevent double domaining when there is no separate registry element
        - BUG LEADING SLASH IF NO REGISTRY THERE: - Found container image: '`/`quay.io/prometheus/pushgateway:v1.11.2@' at path 'image'
- Found container image: '/quay.io/prometheus/pushgateway:v1.11.2@' at path 'image'
- [ ] Retest personal list charts
  - [ ] Retest kafka-operator
    - go run main.go package -r oci://quay.io/strimzi-helm/strimzi-kafka-operator
    - go run main.go deploy -p $TMPDIR/kod-strimzi-kafka-operator-???.kodpkg -n kafka-operator -d myko -s zot --create-secret -r $REGISTRY
  - [ ] Retest ECK operator
  - [ ] Retest RabbitMQ Operator (if non-Bitnami can be found)
  - [ ] Retest OPA
  - [ ] Retest MongoDB (if non-Bitnami can be found)
- [ ] Improve error handling and logging
  - [ ] Add in debug logging support alongside STDOUT info for humans
  - [ ] Add better error handling to invocation of helm upgrade --install as it currently hangs even without --wait, E.g. if one pod doesn't come up right
    - We can now use a cancellable golang function to do this
- [ ] Consider replacing call to Unarchive for chart tgz in package so we can remove the archives module dependency entirely
  - Will mean we drop Windows support, unless they have GNU tar installed (and xz libs)
  - Alternatively, default to tar where it is available, and use the archives version as a backup alternative (extra file size overhead is minimal)
- [ ] Add examples to the help output of all kod commands. E.g. chart locally, chart on OCI url etc.
- [ ] Apply Linux Foundation standards to repo - code of conduct, contributing, security reporting, etc.
- [ ] BUG: Local deployed image sha256 digest doesn't match one downloaded from source repo - do we need to use the --preserve-digests tag in skopeo copy?
- [ ] Ensure deploying a chart from a folder with a subchart still works
  - i.e. ensure helm dependency update is called to populate ./charts in the helm chart
- [ ] How do we know we have succeeded?
  - [ ] Postgres chart
  - [ ] Cloudnative Postgres operator chart (NOT Bitnami version) - https://github.com/cloudnative-pg/charts
  - [ ] Cert-manager chart works (Remote chart, non-standard image location)
    - imageRegistry, imageNamespace and image.name, weirdly, OR image.repository to override ALL THREE
    - go run main.go package -r https://charts.jetstack.io -c cert-manager -f chartfiles/cert-manager/values.yaml
    - go run main.go deploy -p $TMPDIR/kod-cert-manager-???.kodpkg -n cert-manager -d mycm -s zot --create-secret -r $REGISTRY
  - [ ] Nifikop operator works (Remote chart + OCI) `go run main.go deploy -n nifi-operator -p /tmp/kod-nifikop-1.16.0.kodpkg -r  https://REGISTRY:8080/ -d nifikop -f examples/charts/nifikop/basic-values.yaml`
  - [ ] kube-prometheus-stack chart works (*mage image refs, and chart dependencies, and sidecar containers in those dependencies)
- [ ] Update docs with helm charts that work
  - [ ] Don't forget to include notable subcharts in that list, and perhaps invoke those that make sense on their own (E.g. Grafana)
- Note: BREAKING CHANGES
  - Changed Hints file 'hints' element to 'images', as they're only container image hints. Added imagePullSecrets as another type of hint

### v0.3.0  release - usage checks and environment support

- [ ] BUG: Remove duplicate values in root hints file on package
  - E.g. grafana.image in kube-promethus-stack chart
  - E.g. image in postgres chart
- [ ] Provide a --hide-registry-urls option on package so that an internal replica of public container images doesn't end up named 'myinternalregistry.local/docker.io/MYNAME:TAG' and is instead just named for its original location 'docker.io/MYNAME:TAG'
  - Note that the image itself may still include this reference as a separate tag value
- [ ] Improve container detection support
  - [ ] Another convention is global.imageRegistry (2 of top ten) (Will override other registry refs)
  - [ ] top level defaultImageRegistry/Repository/Tag: Strimzi Kafka Operator
    - defaultImageRegistry: quay.io
      defaultImageRepository: strimzi
      defaultImageTag: 0.51.0
    - image.name, (blank image.registry,repository,tag) x13
    - Just a variation on image.repository / global.imageRepository
  - [ ] Allow specification of container images, and links to config elements, for things we cannot guess - kod-hints.yaml
    - Note: include this as CHART-CHARTVER-hints-00x.yaml in the output archive too
  - [ ] Consider reading images required from the deployment YAML templates (helm template), and reverse parse their parameters (Complex)
    - If it works, we should be able to support any helm chart OOTB without a hints file
  - [ ] Improve container version detection by checking `inspect` output RepoTags list for the version, or latest, or first listed (most recent) version before executing `skopeo copy`
  - [ ] Consider reading images required from a commented values file (so it can be used in package -f with no additional hints file specific to kod)
- [ ] Packaging improvements
  - [ ] Allow specification of chart version in package (especially useful for oci repos, and for our documentation to be version matched) 
  - [ ] Check if we should be using alias or name for subchart folder name processing (and sub values mapping) - See cloudnative-pg chart dependencies
  - [ ] Check to see if we have global.imagePullSecrets then can we remove DEPENDENCY.global.imagePullSecrets?
    - [ ] Same check for global.imageRegistry
  - [ ] Handle wildcarded subchart versions. E.g. see Prometheus subchart as all are wildcarded - mv wildcarded folder (with content) to exact version (no content, with hints)
- [ ] Consider splitting deploy into populate and deploy commands so we have a separation of concerns between those who can write to the repo and deploy the package
  - Still make just executing deploy check if it needs to populate, if executed from a .kodpkg
  - Document the different personae of users / actors in invoking kod
  - Have populate command create a mypackage-myversion.kodapp file that describes where locally content resides, and that can be passed to a deployer and used in the deploy command
    - By definition, this should include a helm values file with the necessary image elements overridden
    - Should be another archive format, with an updated kod-package.yaml and a new kod-app.yaml and kod-values.yaml included
      - We may not need a kod-app.yaml file, but instead update the container image references in kod-package.yaml, and change helm reference to an OCI artifact URL
- [ ] Improve provenance on deploy/populate
  - Currently, we lose the link in our yaml files from original container refs to local copy of container refs, with values file being the only place they are stored
  - We need to add kod-package-local.yaml with containers overlay with local{registry,repository,tag,digest} fields
  - Separate file so that the sha256sum of this file is not modified
- [ ] Ensure Istio works with kod (Because I personally really want to use it with Istio!)
- [ ] Add in checks for all required/optional values for each command, and sensible values checks
- [ ] Enforce best practice. E.g. no 'latest' tag versions, and add --allow-latest-tag flag
- [ ] Error if container image being copied matches an existing one but has a different Digest value
- [ ] Check for different versions of the same container image being used, and output a warning to CLI
- [ ] Allow specification of target architecture, or all in the same - default to current platform arch and allow --arch=all and --arch=arm64 etc?
  - If specified, place name of target architecture in the archive file name
  - Add specification of container architecture for deployment, if supported by skopeo (i.e. from 'all' archs archive)
    - We should probably not allow '--arch all', to make the point that you should copy the right file to the right arch, and invoke kod multiple times
  - Add specification of target platform (Windows, Linux, MacOS et al) via --os flag in package
- [ ] Write tests
  - [ ] Modularise other functionality as required
  - [ ] Create scenarios for all current functionality
  - [ ] Ensure expected tests pass
  - [ ] Write tests for failure options (E.g. bad URL, bad folder, strange characters etc) and ensure they are handled

### v0.4.0 release - more complex chart installation use cases

- [ ] Add single helmfile support
  - [ ] Check for helmfile CLI on the path
  - [ ] Basic single helm chart support
  - [ ] Multiple helm charts, multiple namespaces support
  - [ ] Same helm chart, multiple namespaces support (E.g. Istio ingress/egress gateways)
  - [ ] helmfile overrides file support (deploy -f option may need its docs changing)
- [ ] Add Chart dependency chart support for helmfile
  - Will require a deploy:list element in the kod-package file to specify the installable ones that are NOT just dependencies
    - This has the side effect of potentially opening up the possibility to include and invoke multiple helm charts/helmfiles
    - This must be avoided, otherwise semantics like -f to specify override helm chart values files and -n for namespace make no sense
    - Prefer multiple invocations of kod and separate packages for distinct charts
- [ ] Create Istio and Kiali Helmfile example
  - [ ] Create helmfile
  - [ ] Test istio base
  - [ ] Test istiod
  - [ ] Test gateway (ingress only at this point)
    - [ ] Fix expected bug where target version containers are not packaged or identified
  - [ ] Test kiali
  - [ ] Test end to end
  - [ ] Update helmfile documentation how-to pages

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
  - Note: Skopeo doesn't support remote OCI registries, so use regclient instead. See https://regclient.org/ 
  - See https://zotregistry.dev/v2.0.0/articles/workflow/ for an example OCI artifact workflow 
  - [ ] Add a populate command which is basically deploy with --no-install specified 
  - [ ] Store deployed/populated helm charts in a target OCI repository, if the registry supports OCI artifacts
  - [ ] Store a kodpkg in the target OCI registry directly (I.e. `populate --no-unpack`)
  - [ ] Rework the kodpkg format as required such that the hints files and values overrides are within their own artifact
    - This may well actually be the kodinfo package format for a helm chart
  - [ ] Ensure `deploy -c` supports an OCI URL to a kodinfo OCI artifact archive
    - Note: Maintain the old file-system based functionality so that we continue to support non-OCI registries or target environments
- Helm enhancements
  - [ ] Check to see if helm needs to be logged in when fetching packages, or errors due to this on package
  - [ ] Consider adding -o option on deploy to specify where in the container/artifact registry to save helm artifacts to
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

## v0.1.0 (Aka MVP v1.0-alpha1) - 06 Apr 2026 - Provides limited value to deploy some basic helm charts without dependencies

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


## Other dev notes

Some useful links:-

- [Terminal compliant Emojis](https://apps.timwhitlock.info/emoji/tables/unicode)

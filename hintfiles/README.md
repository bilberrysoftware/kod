# Hint files

This folder contains officially supported Helm Chart hints files for common Helm Charts.

A hints file, normally called `kod-hints.yaml` when installed in a `.kodpkg` file, contains information
on Helm values file properties and how they relate to container images included in a helm chart.

Normally `kod` does a very good job at finding references to container images but, occasionally, does need
a pointer when the helm chart author embraces artistic license and voyages far from the well trodden road
of helm conventions. Bless them.

It is our hope in future that helm chart authors either succumb to convention and focus their artistic streak on
their actual application code instead, or provide these hint files themselves as pull requests into this repo folder.

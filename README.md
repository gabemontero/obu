# obu

Stands for OpenShift Build Utilities

This command has a set of verbs that facilitate building images in a OpenShift cluster via the Tekton framework.

The current set of verbs:
* `translate-ist` takes an OpenShift Image Stream Tag reference and produces the preferred image pull reference based on the associated Image Stream specification.

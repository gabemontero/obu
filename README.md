# obu

Stands for OpenShift Build Utilities

This command has a set of verbs that facilitate building images in a OpenShift cluster via the Tekton framework.

The current set of verbs:
* `translate` takes an OpenShift Image Stream Tag reference and produces the preferred image pull reference based on the 
associated Image Stream specification.
* `proxy` interrogates the OpenShift global proxy configuration and produces output easily consumable from command line 
build tools
* `registry` prints contents of either the Docker config file for authentication with the OpenShift internal registry or
the ca.crt contents for HTTPS communication

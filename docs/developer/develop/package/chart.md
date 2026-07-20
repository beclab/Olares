# Olares Application Chart Structure

The Olares Application Chart builds on the standard **Helm Chart** structure and adds extensions to support Olares-specific information. Typically, a standard application chart directory for both `App` and `Middleware` will include the following files:

```
AppName
|-- Chart.yaml                   # Metadata for the chart
|-- OlaresManifest.yaml          # Olares application-specific configuration
|-- templates/                   # Templates for deployment resources
|   |-- deployment.yaml          # Deployment resource definition
|-- owners                       # Required for Market submissions; lists the GitHub accounts permitted to maintain and update this application
|-- crds/                        # OPTIONAL: Custom Resource Definitions
|-- values.yaml                  # OPTIONAL: The default deployment parameters for this chart
|-- values.schema.json           # OPTIONAL: A JSON Schema for imposing a structure on the values.yaml file
|-- README.md                    # OPTIONAL: A human-readable documentation file about the app
```
:::info NOTE
To make the `templates` directory easier to understand, you can split the deployment into several files.
:::
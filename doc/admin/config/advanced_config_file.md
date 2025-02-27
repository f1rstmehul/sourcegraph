# Loading configuration via the file system (advanced)

Some teams require Sourcegraph configuration to be stored in version control as opposed to editing via the Site admin UI.

As of Sourcegraph v3.4+, this is possible for [site configuration](site_config.md), [code host configuration](../external_service/index.md), and global settings.

## Benefits

1. Configuration can be checked into version control (e.g., Git).
1. Configuration is enforced across the entire instance, and edits cannot be made via the web UI (by default).

## Drawbacks

Loading configuration in this manner has two significant drawbacks:

1. You will no longer be able to save configuration edits through the web UI by default (you can use the web UI as scratch space, though).
1. Sourcegraph sometimes performs automatic migrations of configuration when upgrading versions. This process will now be more manual for you (see below).

## Site configuration

Set `SITE_CONFIG_FILE=site.json` on:

- [Docker Compose](../install/docker-compose/index.md) and [Kubernetes](../install/kubernetes/index.md): all `frontend` containers
- [Single-container](../install/docker/index.md): the `sourcegraph/server` container

Where `site.json` is a file that contains the [site configuration](site_config.md), which you would otherwise edit through the in-app site configuration editor.

If you want to _allow_ edits to be made through the web UI (which will be overwritten with what is in the file on a subsequent restart), you may additionally set `SITE_CONFIG_ALLOW_EDITS=true`.

> NOTE: If you do enable this, it is your responsibility to ensure the configuration on your instance and in the file remain in sync.

## Code host configuration

Set `EXTSVC_CONFIG_FILE=extsvc.json` on:

- [Docker Compose](../install/docker-compose/index.md) and [Kubernetes](../install/kubernetes/index.md): all `frontend` containers
- [Single-container](../install/docker/index.md): the `sourcegraph/server` container

Where `extsvc.json` contains a JSON object that specifies _all_ of your code hosts in a single JSONC file:

```jsonc

{
  "GITHUB": [
    {
      // First GitHub code host configuration: literally the JSON object from the code host config editor.
      "authorization": {},
      "url": "https://github.com",
      "token": "...",
      "repositoryQuery": ["affiliated"]
    },
    {
      // Another GitHub code host configuration.
      ...
    },
  ],
  "OTHER": [
    {
      // First "Generic Git host" code host configuration.
      "url": "https://mycodehost.example.com/repos",
      "repos": ["foo"],
    }
  ],
  "PHABRICATOR": [
    {
      // Phabricator code host configuration.
      ...
    },
  ]
}
```

You can find a full list of [valid top-level keys here](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b7ebb9024e3a95109fdedfb8057795b9a7c638bc/-/blob/cmd/frontend/graphqlbackend/schema.graphql#L1104-1110).

If you want to _allow_ edits to be made through the web UI (which will be overwritten with what is in the file on a subsequent restart), you may additionally set `EXTSVC_CONFIG_ALLOW_EDITS=true`. **Note** that if you do enable this, it is your responsibility to ensure the configuration on your instance and in the file remain in sync.

## Global settings

Set `GLOBAL_SETTINGS_FILE=global-settings.json` on:

- [Docker Compose](../install/docker-compose/index.md) and [Kubernetes](../install/kubernetes/index.md): all `frontend` containers
- [Single-container](../install/docker/index.md): the `sourcegraph/server` container

Where `global-settings.json` contains the global settings, which you would otherwise edit through the in-app global settings editor.

If you want to _allow_ edits to be made through the web UI (which will be overwritten with what is in the file on a subsequent restart), you may additionally set `GLOBAL_SETTINGS_ALLOW_EDITS=true`. Note that if you do enable this, it is your responsibility to ensure the global settings on your instance and in the file remain in sync.

## Upgrades and Migrations

As mentioned earlier, when configuration is loaded via the filesystem, Sourcegraph can no longer persist the automatic migrations to configuration it may perform when upgrading.

It will still perform such migrations on the configuration loaded from file, it just cannot persist such migrations **back to file**.

When you upgrade Sourcegraph, you should do the following to ensure your configurations do not become invalid:

1. Upgrade Sourcegraph to the new version
1. Visit each configuration page in the web UI (management console, site configuration, each code host)
1. Copy the (now migrated) configuration from those pages into your JSON files.

It is essential to follow the above steps after **every** Sourcegraph version update, because we only guarantee migrations remain valid across two minor versions. If you fail to apply a migration and later upgrade Sourcegraph twice more, you may effectively "skip" an important migration.

We're planning to improve this by having Sourcegraph notify you as a site admin when you should do the above, since today it is not actually required in most upgrades. See https://github.com/sourcegraph/sourcegraph/issues/4650 for details. In the meantime, we will do our best to communicate when this is needed to you through the changelog.

## Kubernetes ConfigMap

You can load these configuration files via a Kubernetes ConfigMap resource. To do so, create a `base/frontend/sourcegraph-frontend.ConfigMap.yaml` file with contents like this:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    description: Sourcegraph configuration files
  labels:
    deploy: sourcegraph
  name: frontend-config-files
data:
  # IMPORTANT: see https://docs.sourcegraph.com/admin/config/advanced_config_file for details on how this works.

  # Global user settings, see: https://docs.sourcegraph.com/admin/config/advanced_config_file#global-settings
  global-settings.json: |
    {
        "search.scopes": [
            {
              "name": "Test code",
              "value": "file:(test|spec)"
            },
            {
              "name": "Non-test files",
              "value": "-file:(test|spec)"
            }
        ],
        "extensions": {
          "sourcegraph/git-extras": true,
        }
      }

  # Site configuration, see: https://docs.sourcegraph.com/admin/config/advanced_config_file#site-configuration
  site.json: |
    {
      "auth.providers": [
        {
          "allowSignup": true,
          "type": "builtin"
        }
      ],
      "externalURL": "https://sourcegraph.example.com",
      "licenseKey": "..."
      }
    }

  # Code host configuration, see: https://docs.sourcegraph.com/admin/config/advanced_config_file#code-host-configuration
  extsvc.json: |
    {
      "GITHUB": [
        {
          "url": "https://github.com",
          "token": "...",
          "repositoryQuery": [
            "none"
          ],
        }
      ]
    }
```

To have Sourcegraph use this new ConfigMap, add the following environment variables to `base/frontend/sourcegraph-frontend.Deployment.yaml`:

```
        - name: SITE_CONFIG_FILE
          value: /etc/sourcegraph/site.json
        - name : GLOBAL_SETTINGS_FILE
          value: /etc/sourcegraph/global-settings.json
        - name : EXTSVC_CONFIG_FILE
          value: /etc/sourcegraph/extsvc.json
```

And instruct Kubernetes to mount the ConfigMap file we created under `/etc/sourcegraph/` by adding the following in your `sourcegraph-frontend.Deployment.yaml` `volumeMounts` section:

```
        volumeMounts:
        - mountPath: /etc/sourcegraph
          name: config-volume
```

And similarly under the `volume` section:

```
      - name: config-volume
        configMap:
          name: frontend-config-files
          defaultMode: 0644
```

Now upon re-running `kubectl-apply-all.sh` Kubernetes should mount your `ConfigMap` into the container as files on disk and you should see them:

```
$ kubectl exec -it sourcegraph-frontend-57dcb4d7db-6bclj -- ls /mnt/
global-settings.json
extsvc.json
site.json
```

Similarly, because we set the environment variables to use those configuration files, the frontend should have loaded them into the database upon startup. You should now see Sourcegraph configured!

If you encounter any issues, please [contact us](mailto:support@sourcegraph.com).

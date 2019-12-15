# mod

[![CircleCI](https://circleci.com/gh/variantdev/mod.svg?style=svg)](https://circleci.com/gh/variantdev/mod)

`mod` is an universal package manager complements task runners and build tools like `make` and [variant](https://github.com/mumoshu/variant).

It turns any set of files in Git/S3/GCS/HTTP as a reusable module with managed, versioned dependencies.

Think of it as a `vgo`, `npm`, `bundle` alternative, but for any project.

## Getting started

Let's assume you have a `Dockerfile` to build a Docker image containing specific version of `helm`:

`Dockerfile`:

```
FROM alpine:3.9

ARG HELM_VERSION=2.14.2

ADD http://storage.googleapis.com/kubernetes-helm/${HELM_FILE_NAME} /tmp
RUN tar -zxvf /tmp/${HELM_FILE_NAME} -C /tmp \
  && mv /tmp/linux-amd64/helm /bin/helm \
  && rm -rf /tmp/* \
  && /bin/helm init --client-only
```

With `mod`, you can automate the process of occasionally updating the version number in `HELM_VERSION`.

Replace the hard-coded version number `2.14.2` with a template expression `{{ .helm_version }}`, that refers to the latest helm version which is tracked by `mod`:

`Dockerfile.tpl`:

```
FROM alpine:3.9

ARG HELM_VERSION={{ .helm_version }}

ADD http://storage.googleapis.com/kubernetes-helm/${HELM_FILE_NAME} /tmp
RUN tar -zxvf /tmp/${HELM_FILE_NAME} -C /tmp \
  && mv /tmp/linux-amd64/helm /bin/helm \
  && rm -rf /tmp/* \
  && /bin/helm init --client-only
```

Create a `variant.mod` that defines:

- How you want to fetch available version numbers of `helm`(`dependencies.helm`)
- Which file to update/how to update it according to which variable managed by `mod`:

`variant.mod`:

```
provisioners:
  files:
    Dockerfile:
      source: Dockerfile.tpl
      arguments:
        helm_version: "{{ .helm.version }}"

helm:
    releasesFrom:
      githubReleases:
        source: helm/helm
    version: "> 1.0.0"
```

`Dockerfile`(Updated automatically by `mod`:

```
FROM alpine:3.9

ARG HELM_VERSION=2.14.3

ADD http://storage.googleapis.com/kubernetes-helm/${HELM_FILE_NAME} /tmp
RUN tar -zxvf /tmp/${HELM_FILE_NAME} -C /tmp \
  && mv /tmp/linux-amd64/helm /bin/helm \
  && rm -rf /tmp/* \
  && /bin/helm init --client-only
```

Run `mod build` and see `mod` retrieves the latest version number of `helm` that satisfies the semver constraint `> 1.0.0` and updates your `Dockerfile` by rendering `Dockerfile.tpl` accordingly.

## Next steps

- [Learn about use-cases](#use-cases)
- [See examples](#examples)
- [Read API reference](#api-reference)

## Use-cases

- Automate Any-Dependency Updates with GitHub Actions v2
  - Use [mod-action, a GitHub action for running mod](https://github.com/variantdev/mod-action) to periodically run `mod` and submit a pull request to update everything managed via `mod` automatically
- Automating container image updates
- Initializing repository manually created from a [GitHub Repository Template](https://help.github.com/en/articles/creating-a-repository-from-a-template)
  - Configure your CI to run `mod up --build --pull-request --title "Initialize this repository"` in response to "repository created" webhook and submit a PR for initialization
- Create and initialize repository from a [GitHub Repository Template](https://help.github.com/en/articles/creating-a-repository-from-a-template)
  - Run `mod create myorg/mytemplate-repo myorg/mynew-repo --build --pull-request`
- Automatically update container image tags in git-managed CRDs (See [flux#1194](https://github.com/fluxcd/flux/issues/1194) for the motivation

### Boilerplate project generator

`mod` is basically the package manager for any make/[variant](https://github.com/mumoshu/variant)-based project with git support. mod create creates a new project from a template repo by rendering all the [gomplate](https://github.com/hairyhenderson/gomplate)-like template on init.

`mod up` updates dependencies of the project originally created from the template repo, re-rendering required files.

## Examples

Navigate to the following examples to see practical usage of `mod`:

- [examples/eks-k8s-vers](https://github.com/variantdev/mod/blob/master/examples/eks-k8s-vers) for updating your [eksctl] cluster on new K8s release
- [examples/image-tag-in-dockerfile-and-config](https://github.com/variantdev/mod/tree/master/examples/image-tag-in-dockerfile-and-config) for updating your `Dockerfile` and `.circleci/config.yml` on new Golang release

## API Reference

## `regexpReplace` provisioner

`regexpReplace` updates any text file like Dockerfile with regular expressions.

Let's say you want to automate updating the base image of the below Dockerfile:

```
FROM helmfile:0.94.0

RUN echo hello
```

You can write a `variant.mod` file like the below so that `mod` knows where is the image tag to be updated:

```yaml
provisioners:
  regexpReplace:
    Dockerfile:
      from: "(FROM helmfile:)(\\S+)(\\s+)"
      to: "${1}{{.Dependencies.helmfile.version}}${3}"

dependencies:
  helmfile:
    releasesFrom:
      dockerImageTags:
        source: quay.io/roboll/helmfile
    version: "> 0.94.0"
```

### `docker` executable provisioner

Setting `provisioners.executables.NAME.platforms[].docker` allows you to run `mod exec -- NAME $args` where the executable is backed by a docker image which is managed by `mod`.

`variant.mod`:

```
parameters:
  defaults:
    version: "1.12.6"

provisioners:
  executables:
    dockergo:
      platforms:
        # Adds $VARIANT_MOD_PATH/mod/cache/CACHE_KEY/dockergo to $PATH
        # Or its shim at $VARIANT_MOD_PATH/MODULE_NAME/shims
      - docker:
          command: go
          image: golang
          tag: '{{.version}}'
          volume:
          - $PWD:/work
          workdir: /work
```

```console
$ go version
go version go1.12.7 darwin/amd64

$ mod exec -- dockergo version
go version go1.12.6 linux/amd64
```

### Template Functions

The following template functions are available for use within template provisioners:

- `{{ hasKey .Foo.Bar "mykey" }}` returns `true` if `.Foo.Bar` is a `map[string]interface{}` and the value for the key `mykey` is set.
- `{{ trimSpace .Str }}` removes spaces, tabs, and new-lines from `.Str`
``
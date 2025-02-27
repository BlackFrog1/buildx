# buildx imagetools inspect

```
docker buildx imagetools inspect [OPTIONS] NAME
```

<!---MARKER_GEN_START-->
Show details of an image in the registry

### Options

| Name | Type | Default | Description |
| --- | --- | --- | --- |
| [`--builder`](#builder) | `string` |  | Override the configured builder instance |
| [`--format`](#format) | `string` | `{{.Manifest}}` | Format the output using the given Go template |
| [`--raw`](#raw) |  |  | Show original, unformatted JSON manifest |


<!---MARKER_GEN_END-->

## Description

Show details of an image in the registry.

```console
$ docker buildx imagetools inspect alpine
Name:      docker.io/library/alpine:latest
MediaType: application/vnd.docker.distribution.manifest.list.v2+json
Digest:    sha256:21a3deaa0d32a8057914f36584b5288d2e5ecc984380bc0118285c70fa8c9300

Manifests:
  Name:      docker.io/library/alpine:latest@sha256:e7d88de73db3d3fd9b2d63aa7f447a10fd0220b7cbf39803c803f2af9ba256b3
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/amd64

  Name:      docker.io/library/alpine:latest@sha256:e047bc2af17934d38c5a7fa9f46d443f1de3a7675546402592ef805cfa929f9d
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm/v6

  Name:      docker.io/library/alpine:latest@sha256:8483ecd016885d8dba70426fda133c30466f661bb041490d525658f1aac73822
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm/v7

  Name:      docker.io/library/alpine:latest@sha256:c74f1b1166784193ea6c8f9440263b9be6cae07dfe35e32a5df7a31358ac2060
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm64/v8

  Name:      docker.io/library/alpine:latest@sha256:2689e157117d2da668ad4699549e55eba1ceb79cb7862368b30919f0488213f4
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/386

  Name:      docker.io/library/alpine:latest@sha256:2042a492bcdd847a01cd7f119cd48caa180da696ed2aedd085001a78664407d6
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/ppc64le

  Name:      docker.io/library/alpine:latest@sha256:49e322ab6690e73a4909f787bcbdb873631264ff4a108cddfd9f9c249ba1d58e
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/s390x
```

## Examples

### <a name="builder"></a> Override the configured builder instance (--builder)

Same as [`buildx --builder`](buildx.md#builder).

### <a name="format"></a> Format the output (--format)

Format the output using the given Go template. Defaults to `{{.Manifest}}` if
unset. Following fields are available:

* `.Name`: provides the reference of the image
* `.Manifest`: provides the manifest or manifest list
* `.Image`: provides the image config
* `.BuildInfo`: provides [build info from image config](https://github.com/moby/buildkit/blob/master/docs/build-repro.md#image-config)

#### `.Name`

```console
$ docker buildx imagetools inspect alpine --format "{{.Name}}"
Name: docker.io/library/alpine:latest
```

#### `.Manifest`

```console
$ docker buildx imagetools inspect crazymax/loop --format "{{.Manifest}}"
Name:      docker.io/crazymax/loop:latest
MediaType: application/vnd.docker.distribution.manifest.v2+json
Digest:    sha256:08602e7340970e92bde5e0a2e887c1fde4d9ae753d1e05efb4c8ef3b609f97f1
```

```console
$ docker buildx imagetools inspect moby/buildkit:master --format "{{.Manifest}}"
Name:      docker.io/moby/buildkit:master
MediaType: application/vnd.docker.distribution.manifest.list.v2+json
Digest:    sha256:3183f7ce54d1efb44c34b84f428ae10aaf141e553c6b52a7ff44cc7083a05a66

Manifests:
  Name:      docker.io/moby/buildkit:master@sha256:667d28c9fb33820ce686887a717a148e89fa77f9097f9352996bbcce99d352b1
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/amd64

  Name:      docker.io/moby/buildkit:master@sha256:71789527b64ab3d7b3de01d364b449cd7f7a3da758218fbf73b9c9aae05a6775
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm/v7

  Name:      docker.io/moby/buildkit:master@sha256:fb64667e1ce6ab0d05478f3a8402af07b27737598dcf9a510fb1d792b13a66be
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/arm64

  Name:      docker.io/moby/buildkit:master@sha256:1c3ddf95a0788e23f72f25800c05abc4458946685e2b66788c3d978cde6da92b
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/s390x

  Name:      docker.io/moby/buildkit:master@sha256:05bcde6d460a284e5bc88026cd070277e8380355de3126cbc8fe8a452708c6b1
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/ppc64le

  Name:      docker.io/moby/buildkit:master@sha256:c04c57765304ab84f4f9807fff3e11605c3a60e16435c734b02c723680f6bd6e
  MediaType: application/vnd.docker.distribution.manifest.v2+json
  Platform:  linux/riscv64
```

#### `.BuildInfo`

```console
$ docker buildx imagetools inspect crazymax/buildx:buildinfo --format "{{.BuildInfo}}"
Name: docker.io/crazymax/buildx:buildinfo
Frontend: dockerfile.v0
Attrs:
  filename:      Dockerfile
  source:        docker/dockerfile-upstream:master-labs
  build-arg:bar: foo
  build-arg:foo: bar
Sources:
  Type: docker-image
  Ref:  docker.io/docker/buildx-bin:0.6.1@sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0
  Pin:  sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0

  Type: docker-image
  Ref:  docker.io/library/alpine:3.13
  Pin:  sha256:026f721af4cf2843e07bba648e158fb35ecc876d822130633cc49f707f0fc88c

  Type: docker-image
  Ref:  docker.io/moby/buildkit:v0.9.0
  Pin:  sha256:8dc668e7f66db1c044aadbed306020743516a94848793e0f81f94a087ee78cab

  Type: docker-image
  Ref:  docker.io/tonistiigi/xx@sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04
  Pin:  sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04

  Type: http
  Ref:  https://raw.githubusercontent.com/moby/moby/master/README.md
  Pin:  sha256:419455202b0ef97e480d7f8199b26a721a417818bc0e2d106975f74323f25e6c
```

#### JSON output

A `json` go template func is also available if you want to render fields as
JSON bytes:

```console
$ docker buildx imagetools inspect crazymax/loop --format "{{json .Manifest}}"
```
```json
{
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "digest": "sha256:08602e7340970e92bde5e0a2e887c1fde4d9ae753d1e05efb4c8ef3b609f97f1",
  "size": 949
}
```

```console
$ docker buildx imagetools inspect moby/buildkit:master --format "{{json .Manifest}}"
```
```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "digest": "sha256:79d97f205e2799d99a3a8ae2a1ef17acb331e11784262c3faada847dc6972c52",
  "size": 2010,
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:bd1e78f06de26610fadf4eb9d04b1a45a545799d6342701726e952cc0c11c912",
      "size": 1158,
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:d37dcced63ec0965824fca644f0ac9efad8569434ec15b4c83adfcb3dcfc743b",
      "size": 1158,
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v7"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:ce142eb2255e6af46f2809e159fd03081697c7605a3de03b9cbe9a52ddb244bf",
      "size": 1158,
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:f59bfb5062fff76ce464bfa4e25ebaaaac887d6818238e119d68613c456d360c",
      "size": 1158,
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:cc96426e0c50a78105d5637d31356db5dd6ec594f21b24276e534a32da09645c",
      "size": 1159,
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:39f9c1e2878e6c333acb23187d6b205ce82ed934c60da326cb2c698192631478",
      "size": 1158,
      "platform": {
        "architecture": "riscv64",
        "os": "linux"
      }
    }
  ]
}
```

```console
$ docker buildx imagetools inspect crazymax/buildx:buildinfo --format "{{json .BuildInfo}}"
```
```json
{
  "frontend": "dockerfile.v0",
  "attrs": {
    "build-arg:bar": "foo",
    "build-arg:foo": "bar",
    "filename": "Dockerfile",
    "source": "crazymax/dockerfile:buildattrs"
  },
  "sources": [
    {
      "type": "docker-image",
      "ref": "docker.io/docker/buildx-bin:0.6.1@sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0",
      "pin": "sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0"
    },
    {
      "type": "docker-image",
      "ref": "docker.io/library/alpine:3.13@sha256:026f721af4cf2843e07bba648e158fb35ecc876d822130633cc49f707f0fc88c",
      "pin": "sha256:026f721af4cf2843e07bba648e158fb35ecc876d822130633cc49f707f0fc88c"
    },
    {
      "type": "docker-image",
      "ref": "docker.io/moby/buildkit:v0.9.0@sha256:8dc668e7f66db1c044aadbed306020743516a94848793e0f81f94a087ee78cab",
      "pin": "sha256:8dc668e7f66db1c044aadbed306020743516a94848793e0f81f94a087ee78cab"
    },
    {
      "type": "docker-image",
      "ref": "docker.io/tonistiigi/xx@sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04",
      "pin": "sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04"
    },
    {
      "type": "http",
      "ref": "https://raw.githubusercontent.com/moby/moby/master/README.md",
      "pin": "sha256:419455202b0ef97e480d7f8199b26a721a417818bc0e2d106975f74323f25e6c"
    }
  ]
}
```

```console
$ docker buildx imagetools inspect crazymax/buildx:buildinfo --format "{{json .}}"
```
```json
{
  "name": "crazymax/buildx:buildinfo",
  "manifest": {
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "digest": "sha256:899d2c7acbc124d406820857bb51d9089717bbe4e22b97eb4bc5789e99f09f83",
    "size": 2628
  },
  "image": {
    "created": "2022-02-24T12:27:43.627154558Z",
    "architecture": "amd64",
    "os": "linux",
    "config": {
      "Env": [
        "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
        "DOCKER_TLS_CERTDIR=/certs",
        "DOCKER_CLI_EXPERIMENTAL=enabled"
      ],
      "Entrypoint": [
        "docker-entrypoint.sh"
      ],
      "Cmd": [
        "sh"
      ]
    },
    "rootfs": {
      "type": "layers",
      "diff_ids": [
        "sha256:7fcb75871b2101082203959c83514ac8a9f4ecfee77a0fe9aa73bbe56afdf1b4",
        "sha256:d3c0b963ff5684160641f936d6a4aa14efc8ff27b6edac255c07f2d03ff92e82",
        "sha256:3f8d78f13fa9b1f35d3bc3f1351d03a027c38018c37baca73f93eecdea17f244",
        "sha256:8e6eb1137b182ae0c3f5d40ca46341fda2eaeeeb5fa516a9a2bf96171238e2e0",
        "sha256:fde4c869a56b54dd76d7352ddaa813fd96202bda30b9dceb2c2f2ad22fa2e6ce",
        "sha256:52025823edb284321af7846419899234b3c66219bf06061692b709875ed0760f",
        "sha256:50adb5982dbf6126c7cf279ac3181d1e39fc9116b610b947a3dadae6f7e7c5bc",
        "sha256:9801c319e1c66c5d295e78b2d3e80547e73c7e3c63a4b71e97c8ca357224af24",
        "sha256:dfbfac44d5d228c49b42194c8a2f470abd6916d072f612a6fb14318e94fde8ae",
        "sha256:3dfb74e19dedf61568b917c19b0fd3ee4580870027ca0b6054baf239855d1322",
        "sha256:b182e707c23e4f19be73f9022a99d2d1ca7bf1ca8f280d40e4d1c10a6f51550e"
      ]
    },
    "history": [
      {
        "created": "2021-11-12T17:19:58.698676655Z",
        "created_by": "/bin/sh -c #(nop) ADD file:5a707b9d6cb5fff532e4c2141bc35707593f21da5528c9e71ae2ddb6ba4a4eb6 in / "
      },
      {
        "created": "2021-11-12T17:19:58.948920855Z",
        "created_by": "/bin/sh -c #(nop)  CMD [\"/bin/sh\"]",
        "empty_layer": true
      },
      {
        "created": "2022-02-24T12:27:38.285594601Z",
        "created_by": "RUN /bin/sh -c apk --update --no-cache add     bash     ca-certificates     openssh-client   \u0026\u0026 rm -rf /tmp/* /var/cache/apk/* # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:41.061874167Z",
        "created_by": "COPY /opt/docker/ /usr/local/bin/ # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:41.174098947Z",
        "created_by": "COPY /usr/bin/buildctl /usr/local/bin/buildctl # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:41.320343683Z",
        "created_by": "COPY /usr/bin/buildkit* /usr/local/bin/ # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:41.447149933Z",
        "created_by": "COPY /buildx /usr/libexec/docker/cli-plugins/docker-buildx # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.057722191Z",
        "created_by": "COPY /opt/docker-compose /usr/libexec/docker/cli-plugins/docker-compose # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.145224134Z",
        "created_by": "ADD https://raw.githubusercontent.com/moby/moby/master/README.md / # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.422212427Z",
        "created_by": "ENV DOCKER_TLS_CERTDIR=/certs",
        "comment": "buildkit.dockerfile.v0",
        "empty_layer": true
      },
      {
        "created": "2022-02-24T12:27:43.422212427Z",
        "created_by": "ENV DOCKER_CLI_EXPERIMENTAL=enabled",
        "comment": "buildkit.dockerfile.v0",
        "empty_layer": true
      },
      {
        "created": "2022-02-24T12:27:43.422212427Z",
        "created_by": "RUN /bin/sh -c docker --version   \u0026\u0026 buildkitd --version   \u0026\u0026 buildctl --version   \u0026\u0026 docker buildx version   \u0026\u0026 docker compose version   \u0026\u0026 mkdir /certs /certs/client   \u0026\u0026 chmod 1777 /certs /certs/client # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.514320155Z",
        "created_by": "COPY rootfs/modprobe.sh /usr/local/bin/modprobe # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.627154558Z",
        "created_by": "COPY rootfs/docker-entrypoint.sh /usr/local/bin/ # buildkit",
        "comment": "buildkit.dockerfile.v0"
      },
      {
        "created": "2022-02-24T12:27:43.627154558Z",
        "created_by": "ENTRYPOINT [\"docker-entrypoint.sh\"]",
        "comment": "buildkit.dockerfile.v0",
        "empty_layer": true
      },
      {
        "created": "2022-02-24T12:27:43.627154558Z",
        "created_by": "CMD [\"sh\"]",
        "comment": "buildkit.dockerfile.v0",
        "empty_layer": true
      }
    ]
  },
  "buildinfo": {
    "frontend": "dockerfile.v0",
    "attrs": {
      "build-arg:bar": "foo",
      "build-arg:foo": "bar",
      "filename": "Dockerfile",
      "source": "docker/dockerfile-upstream:master-labs"
    },
    "sources": [
      {
        "type": "docker-image",
        "ref": "docker.io/docker/buildx-bin:0.6.1@sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0",
        "pin": "sha256:a652ced4a4141977c7daaed0a074dcd9844a78d7d2615465b12f433ae6dd29f0"
      },
      {
        "type": "docker-image",
        "ref": "docker.io/library/alpine:3.13",
        "pin": "sha256:026f721af4cf2843e07bba648e158fb35ecc876d822130633cc49f707f0fc88c"
      },
      {
        "type": "docker-image",
        "ref": "docker.io/moby/buildkit:v0.9.0",
        "pin": "sha256:8dc668e7f66db1c044aadbed306020743516a94848793e0f81f94a087ee78cab"
      },
      {
        "type": "docker-image",
        "ref": "docker.io/tonistiigi/xx@sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04",
        "pin": "sha256:21a61be4744f6531cb5f33b0e6f40ede41fa3a1b8c82d5946178f80cc84bfc04"
      },
      {
        "type": "http",
        "ref": "https://raw.githubusercontent.com/moby/moby/master/README.md",
        "pin": "sha256:419455202b0ef97e480d7f8199b26a721a417818bc0e2d106975f74323f25e6c"
      }
    ]
  }
}
```

#### Multi-platform

Multi-platform images are supported for `.Image` and `.BuildInfo` fields. If
you want to pick up a specific platform, you can specify it using the `index`
go template function:

```console
$ docker buildx imagetools inspect --format '{{json (index .Image "linux/s390x")}}' moby/buildkit:master
```
```json
{
  "created": "2022-02-25T17:13:27.89891722Z",
  "architecture": "s390x",
  "os": "linux",
  "config": {
    "Env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
    ],
    "Entrypoint": [
      "buildkitd"
    ],
    "Volumes": {
      "/var/lib/buildkit": {}
    }
  },
  "rootfs": {
    "type": "layers",
    "diff_ids": [
      "sha256:41048e32d0684349141cf05f629c5fc3c5915d1f3426b66dbb8953a540e01e1e",
      "sha256:2651209b9208fff6c053bc3c17353cb07874e50f1a9bc96d6afd03aef63de76a",
      "sha256:6741ed7e73039d853fa8902246a4c7e8bf9dd09652fd1b08251bc5f9e8876a7f",
      "sha256:92ac046adeeb65c86ae3f0b458dee04ad4a462e417661c04d77642c66494f69b"
    ]
  },
  "history": [
    {
      "created": "2021-11-24T20:41:23.709681315Z",
      "created_by": "/bin/sh -c #(nop) ADD file:cd24c711a2ef431b3ff94f9a02bfc42f159bc60de1d0eceecafea4e8af02441d in / "
    },
    {
      "created": "2021-11-24T20:41:23.94211262Z",
      "created_by": "/bin/sh -c #(nop)  CMD [\"/bin/sh\"]",
      "empty_layer": true
    },
    {
      "created": "2022-01-26T18:15:21.449825391Z",
      "created_by": "RUN /bin/sh -c apk add --no-cache fuse3 git openssh pigz xz   \u0026\u0026 ln -s fusermount3 /usr/bin/fusermount # buildkit",
      "comment": "buildkit.dockerfile.v0"
    },
    {
      "created": "2022-02-24T00:34:00.924540012Z",
      "created_by": "COPY examples/buildctl-daemonless/buildctl-daemonless.sh /usr/bin/ # buildkit",
      "comment": "buildkit.dockerfile.v0"
    },
    {
      "created": "2022-02-25T17:13:27.89891722Z",
      "created_by": "VOLUME [/var/lib/buildkit]",
      "comment": "buildkit.dockerfile.v0",
      "empty_layer": true
    },
    {
      "created": "2022-02-25T17:13:27.89891722Z",
      "created_by": "COPY / /usr/bin/ # buildkit",
      "comment": "buildkit.dockerfile.v0"
    },
    {
      "created": "2022-02-25T17:13:27.89891722Z",
      "created_by": "ENTRYPOINT [\"buildkitd\"]",
      "comment": "buildkit.dockerfile.v0",
      "empty_layer": true
    }
  ]
}
```

### <a name="raw"></a> Show original, unformatted JSON manifest (--raw)

Use the `--raw` option to print the unformatted JSON manifest bytes.

> `jq` is used here to get a better rendering of the output result.

```console
$ docker buildx imagetools inspect --raw crazymax/loop | jq
```
```json
{
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "digest": "sha256:7ace7d324e79b360b2db8b820d83081863d96d22e734cdf297a8e7fd83f6ceb3",
    "size": 2298
  },
  "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "digest": "sha256:5843afab387455b37944e709ee8c78d7520df80f8d01cf7f861aae63beeddb6b",
      "size": 2811478
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "digest": "sha256:726d3732a87e1c430d67e8969de6b222a889d45e045ebae1a008a37ba38f3b1f",
      "size": 1776812
    },
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "digest": "sha256:5d7cf9b33148a8f220c84f27dd2cfae46aca019a3ea3fbf7274f6d6dbfae8f3b",
      "size": 382855
    }
  ]
}
```

```console
$ docker buildx imagetools inspect --raw moby/buildkit:master | jq
```
```json
{
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:667d28c9fb33820ce686887a717a148e89fa77f9097f9352996bbcce99d352b1",
      "size": 1158,
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:71789527b64ab3d7b3de01d364b449cd7f7a3da758218fbf73b9c9aae05a6775",
      "size": 1158,
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v7"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:fb64667e1ce6ab0d05478f3a8402af07b27737598dcf9a510fb1d792b13a66be",
      "size": 1158,
      "platform": {
        "architecture": "arm64",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:1c3ddf95a0788e23f72f25800c05abc4458946685e2b66788c3d978cde6da92b",
      "size": 1158,
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:05bcde6d460a284e5bc88026cd070277e8380355de3126cbc8fe8a452708c6b1",
      "size": 1159,
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:c04c57765304ab84f4f9807fff3e11605c3a60e16435c734b02c723680f6bd6e",
      "size": 1158,
      "platform": {
        "architecture": "riscv64",
        "os": "linux"
      }
    }
  ]
}
```

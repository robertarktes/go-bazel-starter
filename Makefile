.PHONY: bazel.gazelle bazel.test bazel.build bazel.buildifier

bazel.gazelle:
	bazel run //:gazelle -- update -r .

bazel.test:
	bazel test //...

bazel.build:
	bazel build //...

bazel.buildifier:
	bazel run @buildifier_prebuilt//:buildifier -r .
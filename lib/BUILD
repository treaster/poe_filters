load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["impl.go"],
    importpath = "treaster/applications/poe_filter/lib",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["impl_test.go"],
    embed = [":go_default_library"],
    deps = ["//go/src/github.com/stretchr/testify/require:go_default_library"],
)

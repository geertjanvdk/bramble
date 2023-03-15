package main

import (
	"github.com/golistic/gomake"
)

var (
	targetVendor           = gomake.TargetVendor
	targetCleanupVendor    = gomake.TargetCleanupVendor
	targetDockerBuildXPush = gomake.TargetDockerBuildXPush
)

func main() {
	targetVendor.Name = "vendor-for-docker"
	targetVendor.Flags = map[string]any{"out": "_vendor"}
	targetCleanupVendor.Flags = map[string]any{"out": "_vendor"}

	targetDockerBuildXPush.Flags = map[string]any{
		"registry": "ghcr.io/kelvin-green",
		"image":    "bramble",
	}
	targetDockerBuildXPush.PreTargets = []*gomake.Target{&targetVendor}
	targetDockerBuildXPush.DeferredTargets = []*gomake.Target{&targetCleanupVendor}

	gomake.RegisterTargets(&targetDockerBuildXPush)
	gomake.Make()
}

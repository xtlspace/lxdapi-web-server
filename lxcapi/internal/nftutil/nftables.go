package nftutil

import (
	"lxcapi/pkg/logger"
	"os/exec"
)

func InitNftTable() error {
	exec.Command("nft", "add", "table", "inet", "lxcnat").Run()
	exec.Command("nft", "add", "chain", "inet", "lxcnat", "prerouting", "{", "type", "nat", "hook", "prerouting", "priority", "-100", ";", "}").Run()
	exec.Command("nft", "add", "chain", "inet", "lxcnat", "postrouting", "{", "type", "nat", "hook", "postrouting", "priority", "100", ";", "}").Run()

	exec.Command("nft", "add", "table", "inet", "lxcip").Run()
	exec.Command("nft", "add", "chain", "inet", "lxcip", "prerouting", "{", "type", "nat", "hook", "prerouting", "priority", "-100", ";", "}").Run()
	exec.Command("nft", "add", "chain", "inet", "lxcip", "postrouting", "{", "type", "nat", "hook", "postrouting", "priority", "100", ";", "}").Run()

	logger.OK("nftables表初始化完成: lxcnat, lxcip")
	return nil
}

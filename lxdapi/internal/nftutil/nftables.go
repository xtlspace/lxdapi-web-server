package nftutil

import (
	"lxdapi/pkg/logger"
	"os/exec"
)

func InitNftTable() error {
	exec.Command("nft", "add", "table", "inet", "lxdnat").Run()
	exec.Command("nft", "add", "chain", "inet", "lxdnat", "prerouting", "{", "type", "nat", "hook", "prerouting", "priority", "-100", ";", "}").Run()
	exec.Command("nft", "add", "chain", "inet", "lxdnat", "postrouting", "{", "type", "nat", "hook", "postrouting", "priority", "100", ";", "}").Run()

	exec.Command("nft", "add", "table", "inet", "lxdip").Run()
	exec.Command("nft", "add", "chain", "inet", "lxdip", "prerouting", "{", "type", "nat", "hook", "prerouting", "priority", "-100", ";", "}").Run()
	exec.Command("nft", "add", "chain", "inet", "lxdip", "postrouting", "{", "type", "nat", "hook", "postrouting", "priority", "100", ";", "}").Run()

	logger.OK("nftables表初始化完成: lxdnat, lxdip")
	return nil
}

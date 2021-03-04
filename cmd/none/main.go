package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"encoding/json"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/version"
	"github.com/containernetworking/plugins/pkg/ns"
	bv "github.com/containernetworking/plugins/pkg/utils/buildversion"
)

type NetConf struct {
	types.NetConf
	CNIOutput string `json:"cniOutput,omitempty"`
	AddHooks [][]string `json:"addHooks,omitempty"`
	DelHooks [][]string `json:"delHooks,omitempty"`
	CheckHooks [][]string `json:"checkHooks,omitempty"`
}

func main() {
	skel.PluginMain(cmdAdd, cmdCheck, cmdDel, version.All, bv.BuildString("none"))
}

func outputCmdArgs(fp io.Writer, args *skel.CmdArgs) {
	fmt.Fprintf(fp, `ContainerID: %s
Netns: %s
IfName: %s
Args: %s
Path: %s
StdinData: %s
----------------------
`,
		args.ContainerID,
		args.Netns,
		args.IfName,
		args.Args,
		args.Path,
		string(args.StdinData))
}

func parseConf(data []byte) (*NetConf, error) {
	conf := &NetConf{}
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("failed to parse")
	}

	return conf, nil
}

func getResult(netConf *NetConf) (*current.Result) {
	if netConf.RawPrevResult == nil {
		return &current.Result{}
	}

	version.ParsePrevResult(&netConf.NetConf)
	result, _ := current.NewResultFromResult(netConf.PrevResult)
	return result
}

func executeHooks(netnsName string, hooks [][]string) {
	netns, err := ns.GetNS(netnsName)
	if err != nil {
		return
	}
	defer netns.Close()

	netns.Do(func(_ ns.NetNS) error {
		for _, hookStrs := range hooks {
			hookCmd := hookStrs[0]
			hookArgs := hookStrs[1:]
			//err := exec.Command(hookCmd, hookArgs...).Run()
			output, err := exec.Command(hookCmd, hookArgs...).Output()
			if err != nil {
				fmt.Fprintf(os.Stderr, "OUTPUT: %v", output)
				fmt.Fprintf(os.Stderr, "ERR: %v", err)
			}
		}
		return nil
	})
}

func cmdAdd(args *skel.CmdArgs) error {
	netConf, _ := parseConf(args.StdinData)
	// Output CNI
	if netConf.CNIOutput != "" {
		fp, _ := os.OpenFile(netConf.CNIOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		defer fp.Close()
		fmt.Fprintf(fp, "CmdAdd\n")
		outputCmdArgs(fp, args)
	}
	// call hooks
	if netConf.AddHooks != nil {
		executeHooks(args.Netns, netConf.AddHooks)
	}
	return types.PrintResult(getResult(netConf), netConf.CNIVersion)
}

func cmdDel(args *skel.CmdArgs) error {
	netConf, _ := parseConf(args.StdinData)
	// Output CNI
	if netConf.CNIOutput != "" {
		fp, _ := os.OpenFile(netConf.CNIOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		defer fp.Close()
		fmt.Fprintf(fp, "CmdDel\n")
		outputCmdArgs(fp, args)
	}
	// call hooks
	if netConf.DelHooks != nil {
		executeHooks(args.Netns, netConf.DelHooks)
	}
	return types.PrintResult(&current.Result{}, netConf.CNIVersion)
}

func cmdCheck(args *skel.CmdArgs) error {
	netConf, _ := parseConf(args.StdinData)
	// Output CNI
	if netConf.CNIOutput != "" {
		fp, _ := os.OpenFile(netConf.CNIOutput, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		defer fp.Close()
		fmt.Fprintf(fp, "CmdCheck\n")
		outputCmdArgs(fp, args)
	}
	// call hooks
	if netConf.CheckHooks != nil {
		executeHooks(args.Netns, netConf.CheckHooks)
	}
	return types.PrintResult(&current.Result{}, netConf.CNIVersion)
}

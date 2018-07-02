package main

import (
	"fmt"
	"sync"
	"net/http"
	"encoding/json"
	"os/exec"
)

type VethPair struct {
	SwitchL string `json:"switchL"`
	SwitchR string `json:"switchR"`
}
type Switch struct {
	SwitchName           string `json:"switchName"`
	SwitchControllerIp   string `json:"switchControllerIp"`
	SwitchControllerPort string `json:"switchControllerPort"`
}

type Topology struct {
	Switches  []Switch
	VethPairs []VethPair
}

type VnfDocker struct {
	Name  string `json:"name"`
	Image string `json:"img"`
}

type OVSDockerPort struct {
	VnfName          string `json:"vnfName"`
	VnfIpAddress     string `json:"vnfIpAddress"`
	VnfInterfaceName string `json:"vnfInterfaceName"`
	SwitchName       string `json:"switchName"`
}

func createSwitch(switchName string, switchControllerIp string, switchControllerPort string, wg *sync.WaitGroup) string {
	// Create Switch
	createSwitch := "ovs-vsctl add-br " + switchName
	_, errCreateSwitch := exec.Command("bash", "-c", createSwitch).Output()
	if errCreateSwitch != nil {
		fmt.Printf("%s", errCreateSwitch)
	}

	// Set Controller
	setController := "ovs-vsctl set-controller " + switchName + " tcp:" + switchControllerIp + ":" + switchControllerPort
	_, errSetController := exec.Command("bash", "-c", setController).Output()
	if errSetController != nil {
		fmt.Printf("%s", errSetController)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "Switch " + switchName + " Created"
}

func deleteSwitch(switchName string, wg *sync.WaitGroup) string {
	// Create Switch
	delSwitch := "ovs-vsctl del-br " + switchName
	_, errDelSwitch := exec.Command("bash", "-c", delSwitch).Output()
	if errDelSwitch != nil {
		fmt.Printf("%s", errDelSwitch)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "Switch " + switchName + " Deleted"
}

func createVethPair(switchL string, switchR string, wg *sync.WaitGroup) string {

	// Create Ip Links
	createIpLink := "ip link add " + switchL + "_" + switchR + " type veth peer name " + switchR + "_" + switchL
	_, errcreateIpLink := exec.Command("bash", "-c", createIpLink).Output()
	if errcreateIpLink != nil {
		fmt.Printf("%s", errcreateIpLink)
	}

	// StartUp Interfaces
	startUpInrefaceL := "ifconfig " + switchL + "_" + switchR + " up"
	_, errStartUpInrefaceL := exec.Command("bash", "-c", startUpInrefaceL).Output()
	if errStartUpInrefaceL != nil {
		fmt.Printf("%s", errStartUpInrefaceL)
	}

	startUpInrefaceR := "ifconfig " + switchR + "_" + switchL + " up"
	_, errStartUpInrefaceR := exec.Command("bash", "-c", startUpInrefaceR).Output()
	if errStartUpInrefaceR != nil {
		fmt.Printf("%s", errStartUpInrefaceR)
	}

	// ADD Interfaces to OVS Bridges
	addInterfaceL := "ovs-vsctl add-port " + switchL + " " + switchL + "_" + switchR
	_, erraddInterfaceL := exec.Command("bash", "-c", addInterfaceL).Output()
	if erraddInterfaceL != nil {
		fmt.Printf("%s", erraddInterfaceL)
	}

	addInterfaceR := "ovs-vsctl add-port " + switchR + " " + switchR + "_" + switchL
	_, erraddInterfaceR := exec.Command("bash", "-c", addInterfaceR).Output()
	if erraddInterfaceR != nil {
		fmt.Printf("%s", erraddInterfaceR)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "Veth pair " + switchL + "_" + switchR + " Created"
}

func deleteVethPair(switchL string, switchR string, wg *sync.WaitGroup) string {

	// delete OVS links
	delOVSLinkL := "ovs-vsctl --if-exists del-port " + switchL + " " + switchL + "_" + switchR
	_, errDelOVSLinkL := exec.Command("bash", "-c", delOVSLinkL).Output()
	if errDelOVSLinkL != nil {
		fmt.Printf("%s", errDelOVSLinkL)
	}

	delOVSLinkR := "ovs-vsctl --if-exists del-port " + switchR + " " + switchR + "_" + switchL
	_, errDelOVSLinkR := exec.Command("bash", "-c", delOVSLinkR).Output()
	if errDelOVSLinkR != nil {
		fmt.Printf("%s", errDelOVSLinkR)
	}

	// Delete veth pair Interfaces
	delVethInterface := "ip link delete " + switchL + "_" + switchR
	_, errDelVethInterface := exec.Command("bash", "-c", delVethInterface).Output()
	if errDelVethInterface != nil {
		fmt.Printf("%s", errDelVethInterface)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "Veth pair " + switchL + "_" + switchR + " Deleted"
}

func createVnfDocker(vnfName string, vnfImage string, wg *sync.WaitGroup) string {

	// delete OVS links
	createVnfDocker := "docker run -dit  --name " + vnfName + " --net=none " + vnfImage + " /bin/bash"
	_, errCreateVnfDocker := exec.Command("bash", "-c", createVnfDocker).Output()
	if errCreateVnfDocker != nil {
		fmt.Printf("%s", errCreateVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "VNF " + vnfName + " Created"
}

func deleteVnfDocker(vnfName string, wg *sync.WaitGroup) string {
	// delete Container
	deleteVnfDocker := "docker rm -f " + vnfName
	_, errDeleteVnfDocker := exec.Command("bash", "-c", deleteVnfDocker).Output()
	if errDeleteVnfDocker != nil {
		fmt.Printf("%s", errDeleteVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	return "VNF " + vnfName + " Deleted"
}

func createOVSDockerPort(vnfName string, vnfIpAddress string, vnfInterfaceName string, switchName string, wg *sync.WaitGroup) string {

	createOVSDockerPort := "ovs-docker add-port " + switchName + " " + vnfInterfaceName + " " + vnfName + " --ipaddress=" + vnfIpAddress + "/24"
	_, errCreateOVSDockerPort := exec.Command("bash", "-c", createOVSDockerPort).Output()
	if errCreateOVSDockerPort != nil {
		fmt.Printf("%s", errCreateOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return "ovs Docker port Created"
}

func deleteAllOVSDockerPort(vnfName string, switchName string, wg *sync.WaitGroup) string {

	deleteOVSDockerPort := "ovs-docker del-ports " + switchName + " " + vnfName
	_, errDeleteOVSDockerPort := exec.Command("bash", "-c", deleteOVSDockerPort).Output()
	if errDeleteOVSDockerPort != nil {
		fmt.Printf("%s", errDeleteOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return "ovs Docker port Deleted"
}

func deleteOVSDockerPort(vnfName string, inaterfaceName string, switchName string, wg *sync.WaitGroup) string {

	deleteOVSDockerPort := "ovs-docker del-ports " + switchName + " " + inaterfaceName + " " + vnfName
	_, errDeleteOVSDockerPort := exec.Command("bash", "-c", deleteOVSDockerPort).Output()
	if errDeleteOVSDockerPort != nil {
		fmt.Printf("%s", errDeleteOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	return "ovs Docker port Deleted"
}

// ************************ Controllers Handler **********************************************************************

func createSwitchHandler(w http.ResponseWriter, r *http.Request) {
	swch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swch)
	if err != nil {
		panic(err)
	}
	if len(swch.SwitchName) == 0 || len(swch.SwitchControllerIp) == 0 || len(swch.SwitchControllerPort) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createSwitch(swch.SwitchName, swch.SwitchControllerIp, swch.SwitchControllerPort, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteSwitchHandler(w http.ResponseWriter, r *http.Request) {
	swch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swch)
	if err != nil {
		panic(err)
	}
	if len(swch.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteSwitch(swch.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createVethPairHandler(w http.ResponseWriter, r *http.Request) {
	vetPair := VethPair{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vetPair)
	if err != nil {
		panic(err)
	}
	if len(vetPair.SwitchL) == 0 || len(vetPair.SwitchR) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createVethPair(vetPair.SwitchL, vetPair.SwitchR, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteVethPairHandler(w http.ResponseWriter, r *http.Request) {
	vetPair := VethPair{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vetPair)
	if err != nil {
		panic(err)
	}
	if len(vetPair.SwitchL) == 0 || len(vetPair.SwitchR) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteVethPair(vetPair.SwitchL, vetPair.SwitchR, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createVNFDockerHandler(w http.ResponseWriter, r *http.Request) {
	vnfDocker := VnfDocker{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vnfDocker)
	if err != nil {
		panic(err)
	}
	if len(vnfDocker.Name) == 0 || len(vnfDocker.Image) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createVnfDocker(vnfDocker.Name, vnfDocker.Image, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteVNFDockerHandler(w http.ResponseWriter, r *http.Request) {
	vnfDocker := VnfDocker{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vnfDocker)
	if err != nil {
		panic(err)
	}
	if len(vnfDocker.Name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteVnfDocker(vnfDocker.Name, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 || len(oVSDockerPort.VnfIpAddress) == 0 || len(oVSDockerPort.VnfInterfaceName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createOVSDockerPort(oVSDockerPort.VnfName, oVSDockerPort.VnfIpAddress, oVSDockerPort.VnfInterfaceName, oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteOVSDockerPort(oVSDockerPort.VnfName, oVSDockerPort.VnfInterfaceName, oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteAllOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteAllOVSDockerPort(oVSDockerPort.VnfName,oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func main() {
	fmt.Println(" ")
	fmt.Println("****  XNFV Http Server Agent  ****")
	fmt.Println("****  By AH.GHORAB            ****")
	fmt.Println("****  Summer 2018             ****")
	fmt.Println("------------------------------------")
	fmt.Println("[*] Agent Running at localhost:8000")
	fmt.Println("[*] Valid rest URLs")
	fmt.Println("[#] - /createSwitch")
	fmt.Println("[#] - /deleteSwitch")
	fmt.Println("[#] - /createVethPair")
	fmt.Println("[#] - /deleteVethPair")
	fmt.Println("[#] - /createVNFDocker")
	fmt.Println("[#] - /deleteVNFDocker")
	fmt.Println("[#] - /createOVSDockerPort")
	fmt.Println("[#] - /deleteOVSDockerPort")
	fmt.Println("[#] - /deleteALlOVSDockerPort")
	fmt.Println(" ")

	// Create/Delete Switch
	http.HandleFunc("/createSwitch", createSwitchHandler)
	http.HandleFunc("/deleteSwitch", deleteSwitchHandler)

	// Create/ Delete Veth Pair
	http.HandleFunc("/createVethPair", createVethPairHandler)
	http.HandleFunc("/deleteVethPair", deleteVethPairHandler)

	// Create/Delete VNF Docker
	http.HandleFunc("/createVNFDocker", createVNFDockerHandler)
	http.HandleFunc("/deleteVNFDocker", deleteVNFDockerHandler)

	// Create/Delete OVS-VNF Docker Ports
	http.HandleFunc("/createOVSDockerPort", createOVSDockerPortHandler)
	http.HandleFunc("/deleteOVSDockerPort", deleteOVSDockerPortHandler)
	http.HandleFunc("/deleteALlOVSDockerPort", deleteAllOVSDockerPortHandler)

	http.ListenAndServe(":8000", nil)


}

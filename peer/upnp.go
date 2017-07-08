package peer

import (
	"strconv"

	log "github.com/Sirupsen/logrus"
	upnp "github.com/huin/goupnp"
	ig1 "github.com/huin/goupnp/dcps/internetgateway1"
	ig2 "github.com/huin/goupnp/dcps/internetgateway2"
)

var (
	// DiscoveredDevices is a list of devices that responded to our UPnP discovery
	// broadcast. Note that most of them will only support some UPnP services.
	DiscoveredDevices []upnp.MaybeRootDevice
	// ServiceClients is a list of clients we can use to communicate with UPnP
	// devices on the local network
	ServiceClients []upnp.ServiceClient
	// Forwarding1Clients is a list of clients we can use to access the UPnP services
	// offered by the devices on this network
	Forwarding1Clients []ig1.Layer3Forwarding1
)

// DiscoverDevices will search for devices available for UPnP communication.
func DiscoverDevices() {
	log.Debug("Starting uPnP device discovery...")

	// Get a list of devices that may or may not support UPnP services
	mrd, err := upnp.DiscoverDevices(ig2.URN_WANIPConnection_2)
	if err != nil {
		log.WithError(err).Fatal("An error occurred during device discovery")
	}
	log.Debug("Discovered ", len(mrd), " devices")
	for _, maybeRootDevice := range mrd {
		DiscoveredDevices = append(DiscoveredDevices, maybeRootDevice)
		log.Debug(maybeRootDevice.Location.String())
	}

	// Get a list of services clients that we can use to communicate with the
	// devices on our network using SOAP
	ServiceClients, errors, err := upnp.NewServiceClients(ig2.URN_WANIPConnection_1)
	if err != nil {
		log.WithError(err).Fatal("An error occurred while getting service clients")
	} else if len(errors) > 0 && log.GetLevel() == log.DebugLevel {
		logFields := log.Fields{}
		for i, err := range errors {
			logFields[strconv.Itoa(i+1)] = err
		}
		log.WithFields(logFields).Warning(len(errors), " error(s) occurred while getting service clients")
	}

	if len(ServiceClients) > 0 {
		log.Debug("Service clients:")
		for _, serviceClient := range ServiceClients {
			log.Debug(serviceClient.Location.String())
		}
	}

	log.Debug("Retreiving layer 3 forwarding clients...")
	Forwarding1Clients, err := layer3Forwarding1Clients()
	if err != nil {
		log.WithError(err).Error("An error occurred while retreiving layer 2 forwarding client(s)")
	} else if log.GetLevel() == log.DebugLevel {
		log.Debug("Found ", len(Forwarding1Clients), " layer 3 forwarding client(s)")
	}
}

func layer3Forwarding1Clients() ([]*ig1.Layer3Forwarding1, error) {
	clients, errors, err := ig1.NewLayer3Forwarding1Clients()
	if err != nil {
		return nil, err
	} else if len(errors) > 0 && log.GetLevel() == log.DebugLevel {
		logFields := log.Fields{}
		for i, err := range errors {
			logFields[strconv.Itoa(i+1)] = err
		}
		log.WithFields(logFields).Warn(len(errors), " error(s) occurred retreiving layer 3 service clients")
	}

	return clients, nil
}

func ActionMap() map[*ig1.Layer3Forwarding1]string {
	actionMap := make(map[*ig1.Layer3Forwarding1]string)

	for _, client := range Forwarding1Clients {
		actionMap[&client] = client.Service.String()
	}

	return actionMap
}

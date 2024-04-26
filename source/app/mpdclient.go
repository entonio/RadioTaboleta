package app

import (
	"log"

	"github.com/fhs/gompd/v2/mpd"
)

type MpdCommand string

const (
	Mp_CurrentSong MpdCommand = "currentsong"
	Mp_AddId       MpdCommand = "addid"
	Mp_Clear       MpdCommand = "clear"
	Mp_Delete      MpdCommand = "delete %d:%d"
	Mp_Pause       MpdCommand = "pause %d"
	Mp_Play        MpdCommand = "play"
	Mp_Stop        MpdCommand = "stop"
	Mp_SetVol      MpdCommand = "setvol %d"
	Mp_Status      MpdCommand = "status"

	Mx_Init        MpdCommand = "Mx_Init"
	Mx_PlayAddress MpdCommand = "Mx_PlayAddress"
)

type MpdState string

const (
	Ms_Paused  MpdState = "pause"
	Ms_Playing MpdState = "play"
	Ms_Stopped MpdState = "stop"
)

//type Mpd mpd.Client

type MpdClient struct {
	address string
}

type MpdStatus struct {
	OK bool

	BitRate int
	State   MpdState
	Volume  int

	CurrentAddress string
	CurrentTitle   string
	CurrentName    string

	Elapsed string

	Error        error
	ErrorAddress string

	Version int
}

var mpdClient *MpdClient

func setupMpdClient() {
	for _, radio := range RADIOS {
		if radio.Predefined {
			SETTINGS.Predefined = radio
			break
		}
	}

	mpdClient = NewMpdClient(SETTINGS.Mpd)
}

func NewMpdClient(address string) *MpdClient {
	return &MpdClient{
		address: address,
	}
}

func (self *MpdClient) foundError(err error, onComplete func(result MpdStatus)) bool {
	if err == nil {
		return false
	}
	log.Println(err)
	onComplete(MpdStatus{
		Error: err,
	})
	return true
}

func (self *MpdClient) Dial(command MpdCommand, parameters []any, onComplete func(result MpdStatus)) {
	go func() {
		self.DialAsync(command, parameters, onComplete)
	}()
}

func (self *MpdClient) DialAsync(command MpdCommand, parameters []any, onComplete func(result MpdStatus)) {

	log.Printf("DIAL %s %+v\n", command, parameters)

	connection, err := mpd.Dial("tcp", self.address)
	if self.foundError(err, onComplete) {
		return
	}
	defer connection.Close()

	switch command {
	case Mp_Status:
		break

	case Mp_CurrentSong:
		break

	case Mx_Init:
		settings := parameters[0].(Settings)
		if settings.Volume > 0 {
			connection.SetVolume(settings.Volume)
		}
		if settings.Radio == "predefined" && len(settings.Predefined.Address) > 0 {
			connection.AddID(settings.Predefined.Address, 0)
		}
		if settings.Playback == "start" {
			connection.Play(0)
		}

	case Mx_PlayAddress:
		address := parameters[0].(string)
		id, err := connection.AddID(address, 0)
		if self.foundError(err, onComplete) {
			return
		}
		err = connection.PlayID(id)
		if err != nil {
			status := MpdStatus{
				Error:        err,
				ErrorAddress: address,
			}
			log.Println(err)
			currentSong, err := connection.CurrentSong()
			if err == nil {
				// selecting the current address results in error, but that's not a problem with the address
				if status.ErrorAddress == currentSong["file"] {
					status.ErrorAddress = ""
				}
			}
			onComplete(status)
			return
		}
		//connection.MoveID(id, 1)
		connection.Command(string(Mp_Delete), 1, 10).OK()

	default:
		err = connection.Command(string(command), parameters...).OK()
		if self.foundError(err, onComplete) {
			return
		}
	}

	status, err := connection.Status()
	log.Println(status)
	if self.foundError(err, onComplete) {
		return
	}

	var currentAddress string
	var currentName string
	var currentTitle string
	currentSong, err := connection.CurrentSong()
	if err == nil {
		currentAddress = currentSong["file"]
		currentName = currentSong["Name"]
		currentTitle = currentSong["Title"]
	}

	onComplete(MpdStatus{
		OK:             true,
		BitRate:        AsInt(status["bitrate"], -1),
		State:          MpdState(status["state"]),
		Volume:         AsInt(status["volume"], -1),
		CurrentAddress: currentAddress,
		CurrentName:    currentName,
		CurrentTitle:   currentTitle,
		Elapsed:        status["elapsed"],
	})
}

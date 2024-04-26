package app

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"

	"github.com/getlantern/systray"
)

var queryStatusCh = make(chan int)

var latestStatus MpdStatus
var displayedStatus MpdStatus

var displayedName string
var displayedTitle string

var mRadios []*systray.MenuItem

var volStep = 10
var mVols1 = make(map[int]*systray.MenuItem)
var mVols2 = make(map[int]*systray.MenuItem)

var mLouder *systray.MenuItem
var mSetVol *systray.MenuItem
var mLower *systray.MenuItem
var mZapOn *systray.MenuItem
var mZapOff *systray.MenuItem
var mResume *systray.MenuItem
var mPause *systray.MenuItem
var mServer *systray.MenuItem
var mStatus *systray.MenuItem
var mName *systray.MenuItem
var mTitle *systray.MenuItem

var updateUICh = make(chan int)

func syncUI(status MpdStatus) {
	latestStatus = status
	updateUICh <- 1
}

func forceSyncUI() {
	newStatus := latestStatus
	newStatus.Version += 1
	syncUI(newStatus)
}

func buildUI() {
	log.Printf("MENU start\n")
	icon := readIconFromConfig()
	//systray.SetTemplateIcon(icon, icon)
	systray.SetIcon(icon)
	//systray.SetTooltip("RádioTaboleta")
	setMenuTitle("RádioTaboleta", false)

	mLower = systray.AddMenuItem("Volume -", "Diminuir o volume")
	mLower.Hide()

	mSetVol = systray.AddMenuItem("Volume %", "Definir o volume")
	mSetVol.Hide()
	mSetVolClickedCh := make(chan int)
	mSetVolListen := func(mSetVolAtIndex *systray.MenuItem, vol int) {
		go func() {
			for {
				<-mSetVolAtIndex.ClickedCh
				mSetVolClickedCh <- vol
			}
		}()
	}

	for _, v1 := range ints(0, volStep, 90) {
		mVols1[v1] = mSetVol.AddSubMenuItem(fmt.Sprintf("%d-%d", v1, v1+volStep-1), "")
		for _, v2 := range ints(v1, 1, v1+volStep-1) {
			mVols2[v2] = mVols1[v1].AddSubMenuItem(fmt.Sprintf("%d", v2), "")
			mSetVolListen(mVols2[v2], v2)
		}
	}

	mLouder = systray.AddMenuItem("Volume +", "Aumentar o volume")
	mLouder.Hide()

	systray.AddSeparator()

	mServer = systray.AddMenuItem("🔄 "+SETTINGS.Mpd, "")
	if len(SETTINGS.Restart) == 0 || !strings.HasPrefix(SETTINGS.Mpd, "localhost:") {
		mServer.Disable()
	}

	mStatus = systray.AddMenuItem("A iniciar...", "")
	mStatus.Disable()

	mName = systray.AddMenuItem("", "")
	mName.Disable()
	mName.Hide()

	mTitle = systray.AddMenuItem("", "")
	mTitle.Hide()

	systray.AddSeparator()

	mZapOn = systray.AddMenuItem("Zapping", "Percorrer as estações")
	mZapOn.Hide()

	mZapOff = systray.AddMenuItem("Zapping", "Parar de percorrer as estações")
	mZapOff.Check()
	mZapOff.Hide()

	mResume = systray.AddMenuItem("Retomar", "Iniciar o rádio")
	mResume.Hide()

	mPause = systray.AddMenuItem("Pausa", "Parar o rádio")
	mPause.Hide()

	systray.AddSeparator()

	mClose := systray.AddMenuItem("Sair", "Fechar")
	systray.AddSeparator()

	mRadioClickedCh := make(chan int)
	mRadioListen := func(mRadioAtIndex *systray.MenuItem, i int) {
		go func() {
			for {
				<-mRadioAtIndex.ClickedCh
				mRadioClickedCh <- i
			}
		}()
	}

	const (
		separator = iota
		item
	)
	latest := separator
	for i, radio := range RADIOS {
		if radio.Separator != nil {
			if latest != separator {
				systray.AddSeparator()
				latest = separator
			}
			if len(*radio.Separator) > 0 {
				systray.AddMenuItem(*radio.Separator, "").Disable()
				latest = item
			}
		}
		mRadioAtIndex := systray.AddMenuItem(radio.Name, "Mudar para "+radio.Name)
		latest = item
		mRadioListen(mRadioAtIndex, i)
		mRadios = append(mRadios, mRadioAtIndex)
	}

	go func() {
		for {
			select {
			case i := <-queryStatusCh:
				done := func(status MpdStatus) {
					reloadStreamIfBroken(status)
					syncUI(status)
					time.AfterFunc(QUERY_STATUS_INTERVAL, func() { queryStatusCh <- 1 })
				}
				if i == 0 {
					mpdClient.Dial(Mx_Init, A(SETTINGS), done)
				} else {
					mpdClient.Dial(Mp_Status, A(), done)
				}

			case i := <-mRadioClickedCh:
				stopZapping()
				playIndex(i)

			case <-mTitle.ClickedCh:
				if len(displayedTitle) > 0 {
					clipboard.WriteAll(displayedTitle)
					split := strings.SplitN(displayedTitle, " - ", 2)
					if len(split) == 2 {
						musicas.Add(split[0], split[1])
					}
				}

			case <-mZapOn.ClickedCh:
				startZapping()

			case <-mZapOff.ClickedCh:
				stopZapping()

			case <-mLouder.ClickedCh:
				mpdClient.Dial(Mp_SetVol, A(volume(+10.0/3)), syncUI)

			case <-mLower.ClickedCh:
				mpdClient.Dial(Mp_SetVol, A(volume(-10.0/3)), syncUI)

			case v := <-mSetVolClickedCh:
				mpdClient.Dial(Mp_SetVol, A(v), syncUI)

			case <-mServer.ClickedCh:
				restartServer()

			case <-mResume.ClickedCh:
				stopZapping()
				mpdClient.Dial(Mp_Play, A(), syncUI)

			case <-mPause.ClickedCh:
				stopZapping()
				mpdClient.Dial(Mp_Stop, A(), syncUI)

			case <-mClose.ClickedCh:
				systray.Quit()
			}
		}
	}()

	log.Printf("MENU wait\n")

	queryStatusCh <- 0

	for {
		<-updateUICh
		updateUI()
	}

	log.Printf("MENU done\n")
}

func reloadStreamIfBroken(newStatus MpdStatus) {
	if zapping != nil {
		return
	}
	if len(newStatus.Elapsed) == 0 {
		return
	}
	if newStatus.Elapsed != latestStatus.Elapsed {
		return
	}
	mpdClient.Dial(Mp_Stop, A(), func(result MpdStatus) {
		mpdClient.Dial(Mp_Play, A(), syncUI)
	})
}

// do all UI handling here
func updateUI() {

	status := latestStatus
	if status == displayedStatus {
		log.Printf("SKIP\n")
		return
	}

	//runtime.LockOSThread()
	//defer runtime.UnlockOSThread()

	log.Printf("STAT %#v\n", status)
	if status.OK {
		if len(status.CurrentAddress) > 0 {
			for i, mRadio := range mRadios {
				if status.CurrentAddress == RADIOS[i].Address {
					mRadio.Check()
					radioName := RADIOS[i].Name
					mRadio.SetTitle(radioName)
					displayedName = radioName
					setMenuTitle(displayedName, true)
				} else {
					mRadio.Uncheck()
				}
			}
			if len(status.CurrentTitle) > 0 {
				statusTitle := status.CurrentTitle

				xmlTitle := readTitleFromXml(statusTitle)
				if len(xmlTitle) > 0 {
					displayedTitle = xmlTitle
				} else {
					displayedTitle = statusTitle
				}

				if SETTINGS.Trim != nil {
					displayedTitle = strings.TrimSpace(SETTINGS.Trim.ReplaceAllString(displayedTitle, ""))
				}

				// 📋📝✏️✍️🖋️🖊️
				mTitle.SetTitle("📋 " + displayedTitle)
				mTitle.Show()
			} else {
				displayedTitle = ""
				mTitle.Hide()
			}
			if len(status.CurrentName) > 0 {
				mName.SetTitle(status.CurrentName)
				mName.Show()
			} else {
				mName.Hide()
			}
		}
		currentVolume = status.Volume
		if status.State == Ms_Playing {
			mResume.Hide()
			mPause.Show()
			if zapping == nil {
				mZapOn.Show()
				mZapOff.Hide()
			} else {
				mZapOn.Hide()
				mZapOff.Show()
			}
			mStatus.SetTitle(fmt.Sprintf("Vol %d%%, %d kbps", currentVolume, status.BitRate))
		} else {
			mResume.Show()
			mPause.Hide()
			mZapOn.Hide()
			mZapOff.Hide()
			mStatus.SetTitle(fmt.Sprintf("Vol %d%%, em pausa", currentVolume))
			if len(displayedName) > 0 {
				setMenuTitle(displayedName, false)
			} else {
				setMenuTitle("Em pausa", false)
			}
		}
		mSetVol.Show()
		mLouder.Show()
		mLower.Show()
		v1 := currentVolume / volStep * volStep
		v2 := currentVolume
		/*
				log.Printf("v1s: %+v\n", mVols1)
				log.Printf("v2s: %+v\n", mVols2)
				log.Printf("v1: %v\n", v1)
				log.Printf("v2: %v\n", v2)
			for k, v := range mVols1 {
				if k == v1 && !v.Checked() {
					v.Check()
				} else if v.Checked() {
					v.Uncheck()
				}
			}
			for k, v := range mVols2 {
				if k == v2 && !v.Checked() {
					v.Check()
				} else if v.Checked() {
					v.Uncheck()
				}
			}
		*/
		for k, v := range mVols1 {
			if k == v1 {
				if !v.Checked() {
					v.Check()
				}
			} else {
				if v.Checked() {
					v.Uncheck()
				}
			}
		}
		for k, v := range mVols2 {
			if k == v2 {
				if !v.Checked() {
					v.Check()
				}
			} else {
				if v.Checked() {
					v.Uncheck()
				}
			}
		}
		/*
			for _, v := range mVols1 {
				v.Uncheck()
			}
			for _, v := range mVols2 {
				v.Uncheck()
			}
			mVols1[v1].Check()
			mVols2[v2].Check()
		*/
	} else {
		mStatus.SetTitle(fmt.Sprintf("Não contactável"))
		if len(status.ErrorAddress) > 0 {
			for i, mRadio := range mRadios {
				if status.ErrorAddress == RADIOS[i].Address {
					radioName := RADIOS[i].Name
					mRadio.SetTitle(radioName + " ❌")
					break
				}
			}
		}
	}

	displayedStatus = status
	/*
		time.AfterFunc(QUERY_STATUS_INTERVAL, func() {
			queryStatusCh <- 1
		})
	*/
}

func setMenuTitle(title string, playing bool) {

	var icon string
	if playing {
		icon = "🔊"
	} else {
		icon = "🔇"
	}

	envelop := func(s string) string {
		return " " + s + " " + icon
	}

	glyphs := func(s string) int {
		return len([]rune(s))
	}

	sides := glyphs(envelop(""))

	maxWidth := LARGEST_TITLE_LENGTH + sides
	// any icons appearing to the left of the notch are hidden!
	if runtime.GOARCH == "arm64" {
		maxWidth = 20
	}

	pad := func(s string, width int) string {
		var left string
		var right string
		d := (width - glyphs(title)) / 2
		if d > 0 {
			left = strings.Repeat("-", d-1) + " "
			right = " " + strings.Repeat("-", d-1)
		}
		return left + s + right
	}

	title = strings.TrimSpace(title)

	displayed := envelop(pad(title, maxWidth-sides))
	extra := glyphs(displayed) - maxWidth
	log.Printf("MENU [%s] %d - %d = %d\n", displayed, glyphs(displayed), maxWidth, extra)
	if extra > 0 {
		displayed = envelop(pad(title[0:glyphs(title)-extra-1]+"·", maxWidth-sides))
		log.Printf("MENU [%s] %d - %d = %d\n", displayed, glyphs(displayed))
	}
	displayed = strings.ReplaceAll(displayed, "--", "—")
	displayed = strings.ReplaceAll(displayed, "-", "")

	systray.SetTitle(displayed)
}

func onStopSystray() {
}

func restartServer() {
	address := latestStatus.CurrentAddress
	log.Printf("Will restart server\n")
	cmd, err := exec.Command(SETTINGS.Restart[0], SETTINGS.Restart[1:]...).Output()
	log.Printf("cmd: %s\n", cmd)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Did restart server\n")
	}
	time.Sleep(QUERY_STATUS_INTERVAL / 2)
	log.Printf("Will set volume\n")
	mpdClient.Dial(Mp_SetVol, A(volume(0)), func(result MpdStatus) {
		log.Printf("Did set volume\n")
		syncUI(result)
		if len(address) > 0 {
			log.Printf("Will start playing %s\n", address)
			mpdClient.Dial(Mx_PlayAddress, A(address), func(result MpdStatus) {
				log.Printf("Did start playing: %v\n", result)
				syncUI(result)
			})
		}
	})
}

var currentVolume int

func volume(offset float32) int {
	if currentVolume < 0 {
		return SETTINGS.Volume
	}
	newValue := int(float32(currentVolume) + offset)
	if newValue < 0 {
		return 0
	}
	if newValue > 100 {
		return 100
	}
	return newValue
}

var zapping *time.Ticker
var zappingIndex = -1

func startZapping() {
	if zapping != nil {
		return
	}

	playNext()
	forceSyncUI()

	zapping = time.NewTicker(SETTINGS.Zapping)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-zapping.C:
				playNext()
			}
		}
	}()
}

func stopZapping() {
	if zapping == nil {
		return
	}
	zapping.Stop()
	zapping = nil
	zappingIndex = -1
	forceSyncUI()
}

func playNext() {
	if zappingIndex < 0 {
		status := latestStatus
		if len(status.CurrentAddress) > 0 {
			for i, _ := range mRadios {
				if status.CurrentAddress == RADIOS[i].Address {
					zappingIndex = i
					break
				}
			}
		}
	}
	zappingIndex += 1
	if zappingIndex >= len(RADIOS) {
		zappingIndex = 0
	}
	playIndex(zappingIndex)
}

func playIndex(i int) {
	mpdClient.Dial(Mx_PlayAddress, A(RADIOS[i].Address), syncUI)
}

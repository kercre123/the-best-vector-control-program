package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

var (
	ctx context.Context
	v   *vector.Vector
	win fyne.Window
)

func blankGrayImage() image.Image {
	rect := image.Rect(0, 0, 640, 480)
	img := image.NewRGBA(rect)
	gray := color.Gray{Y: 128}
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.Set(x, y, gray)
		}
	}
	return img
}

const privateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAp8wMPSe9LPHUKdnGmQd4uPpS1Ip6osRLUIq+KbMGw64FUIJh
/arLzQ3WkFMoV92iGOJqvCg2ddjLtQS9XVh6IjKjTkiOf8hUK5p5sjuVINz/z4cN
N+mFu0mklOTjLYgQUbikHcdQHNohCinQLGqCZA9pwnpeg86l9x9O3bV8Vw7vCadV
2XOnAJUeCCPCP79ifcLV7ts0IE+9j7774ZtuV/iJSZD7r6sbeWNNjXim/3uAqUos
CsTuwAiJl4tofjhLk1dlz732T8UGbYOE8si+BvH7OsOKjYkuw3RBAxqVH0qryl6V
uNyIYp0ehtqu6Qp7BWknUeWqzapvxSluK5Ve7QIDAQABAoIBAFa59yVwsa1WPJN4
9NXJb9MjxsYF4QbZsBer7ke9OWTQP/zxttYWfgm4+kpUQMjRS+PSutoPar6UVA12
qq1hepbMV22xwL04/JAg4n+FnjmDIFDR+7oHX9CCaqdueiDhb5Xdei3OA5E2CNeo
7ujWEBjJgp87AjjcCRnmO6wKDn8r6YfR0tG50Yqf1XjBksRGWy+4hTsSDRT3xUgw
fgByH5YLuc1ZI12eZSJhWn1K7jnJAGEZ2RFag5yCbhvWgBQNUVcIgWJt3GYEAFix
5gsI7UAw5ylIH9F6kl8vpvFTx326AkcBCjMLY9psAHrgRCcG88QJeEHsW0ET2akE
IknGG10CgYEA0rcUyEYYfu1MQjDDADekRPP/TwClJEI6sPE75Big0MzV78PRfwrB
cLHmfEqfHrR9TNwLz/Unbq6aWF85LCof3qFrU4IbXyz+JwL2/8seZ9fsWrbrI6En
ZR9PIftqPtbinxbap6t+ABT1RYkJ/HTI2pJE+/fQTS3GjEw6XJzL2YMCgYEAy9u9
sdz+MB7xdiI4j/xxjHQZJeDcvLAeJZUW44Jjv+Bn2f0TeDWuYwdYkHy88/nzXvpO
3zNP93iZeF+Igfm9pdQnXZfN0Fcvok6yeBK7HrmNaZmMeDu96Ky5BfYkXvKX2/Y/
Ntq2p8J5p/Nq9qT+qujdaZf51PbJg64oBUrbKs8CgYEAyqEAPS8a80Pip2wYuSbI
sv4oL6KhK+L8aZcxTsFYNDImMLEPzqlbJ7ILwM5Jgc9zBuw797j6OHdzOTQo2I2R
pBd6DA37oGS16nHxcD21eYqsYPex2stoBNg80qLgopklyHLDxaUmP5Hn4vxLPBhZ
5cXuzJacGvvACL5tCQ5HAV0CgYB++idFA0bcyFlUYOpkXTSI7MPBQTec3AJbHGs+
WLgzCt8E+8rFxIITsr6qeNflC9pYXYcFJdv4ZAkL3k2Tz/Adu3CtrmGHFNdZvLUT
b29YKvF3RiolteiLZhJ1MSTkcyy92Lr1OvQsuEi4oTkN2iW6ZQOMwxndWb6ZI8BP
05mCJwKBgQDQ4OGBMXL6jIWf7c+lmK/sHg4uzY9JGOGRIUVVqK5J2yST4FlK00f2
k+nITtKMig4m+8w2FQ7cFjJ2kzh+DXX3/0fl69iJRCBnJKDY7tH3d2kWCLLkNhAu
cPNasU815tacMMSMjlCrJq2woLrHM8ToOKpQIbIkpoXcpo0Zh+ceUQ==
-----END RSA PRIVATE KEY-----`

func fetchToken(ip string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	conf := &ssh.ClientConfig{
		User:            "root",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", ip+":22", conf)
	if err != nil {
		return "", err
	}
	sess, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()
	b, err := sess.Output("cat /run/vic-cloud/perRuntimeToken")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func main() {
	a := app.New()
	win = a.NewWindow("stupid")
	ipEntry := widget.NewEntry()
	ipEntry.SetPlaceHolder("enter robo ip")
	status := widget.NewLabel("")
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "IP", Widget: ipEntry},
		},
		OnSubmit: func() {
			status.SetText("connecting...")
			go func() {
				ip := strings.TrimSpace(ipEntry.Text)
				tok, err := fetchToken(ip)
				if err != nil {
					status.SetText("err: " + err.Error())
					return
				}
				ctx = context.Background()
				v, err = vector.New(
					// esn doesn't matter in WireOS!
					vector.WithSerialNo("00601b50"),
					vector.WithTarget(ip+":443"),
					vector.WithToken(tok),
				)
				if err != nil {
					status.SetText("new vector err: " + err.Error())
					return
				}
				fyne.Do(func() { win.SetContent(buildMainUI()) })
			}()
		},
	}
	setup := container.NewVBox(form, status)
	win.SetContent(setup)
	win.Resize(fyne.NewSize(400, 100))
	win.ShowAndRun()
}

func buildMainUI() fyne.CanvasObject {
	cfc, err := v.Conn.CameraFeed(ctx, &vectorpb.CameraFeedRequest{})
	if err != nil {
		return widget.NewLabel("camera feed err: " + err.Error())
	}
	imgW := canvas.NewImageFromImage(blankGrayImage())
	imgW.SetMinSize(fyne.NewSize(640, 480))
	imgW.FillMode = canvas.ImageFillContain
	entry := newEnterEntry()
	entry.SetPlaceHolder("click here and start WASDing")

	go func() {
		go assumeBehaviorControl(v, "high")
		for {
			resp, err := cfc.Recv()
			if err != nil {
				fmt.Println("recv err:", err)
				os.Exit(1)
				return
			}
			img, _, err := image.Decode(bytes.NewReader(resp.Data))
			if err == nil {
				imgW.Image = img
				fyne.Do(func() { imgW.Refresh() })
			}
		}
	}()
	return container.NewVBox(imgW, entry)
}

type enterEntry struct {
	widget.Entry
}

func newEnterEntry() *enterEntry {
	entry := &enterEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *enterEntry) KeyUp(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyW:
		W = false
	case fyne.KeyA:
		A = false
	case fyne.KeyS:
		S = false
	case fyne.KeyD:
		D = false
	case fyne.KeyR:
		R = false
	case fyne.KeyF:
		F = false
	case fyne.KeyT:
		T = false
	case fyne.KeyG:
		G = false
	case "LeftShift":
		SHIFT = false
	default:
		e.Entry.KeyDown(key)
	}
	e.SetText("")
	stateChecker()
}

func (e *enterEntry) KeyDown(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyW:
		W = true
	case fyne.KeyA:
		A = true
	case fyne.KeyS:
		S = true
	case fyne.KeyD:
		D = true
	case fyne.KeyR:
		R = true
	case fyne.KeyF:
		F = true
	case fyne.KeyT:
		T = true
	case fyne.KeyG:
		G = true
	case "LeftShift":
		SHIFT = true
	default:
		e.Entry.KeyDown(key)
	}
	e.SetText("")
	stateChecker()
}

var W, A, S, D, R, F, T, G, SHIFT bool

func stateChecker() {
	var mult float32 = 1
	if SHIFT {
		mult = 2
	}
	switch {
	case W && A && D:
		setWheelMotor(v, 120*mult, 120*mult)
	case W && A:
		setWheelMotor(v, 50*mult, 150*mult)
	case W && D:
		setWheelMotor(v, 150*mult, 50*mult)
	case W && S:
		setWheelMotor(v, 0, 0)
	case S && A && D:
		setWheelMotor(v, -120*mult, -120*mult)
	case S && A:
		setWheelMotor(v, -150*mult, -50*mult)
	case S && D:
		setWheelMotor(v, -50*mult, -150*mult)
	case W:
		setWheelMotor(v, 120*mult, 120*mult)
	case S:
		setWheelMotor(v, -120*mult, -120*mult)
	case D:
		setWheelMotor(v, 120*mult, -120*mult)
	case A:
		setWheelMotor(v, -120*mult, 120*mult)
	default:
		setWheelMotor(v, 0, 0)
	}

	switch {
	case F && R:
		setHeadMotor(v, 0)
	case R:
		setHeadMotor(v, 3*mult)
	case F:
		setHeadMotor(v, -3*mult)
	default:
		setHeadMotor(v, 0)
	}

	switch {
	case T && G:
		setLift(v, 0)
	case T:
		setLift(v, 3*mult)
	case G:
		setLift(v, -3*mult)
	default:
		setLift(v, 0)
	}
}

func setHeadMotor(v *vector.Vector, speed float32) {
	v.Conn.MoveHead(ctx, &vectorpb.MoveHeadRequest{
		SpeedRadPerSec: speed,
	})
}

func setLift(v *vector.Vector, speed float32) {
	v.Conn.MoveLift(ctx, &vectorpb.MoveLiftRequest{
		SpeedRadPerSec: speed,
	})
}
func setWheelMotor(v *vector.Vector, speedLeft, speedRight float32) {
	v.Conn.DriveWheels(ctx,
		&vectorpb.DriveWheelsRequest{
			LeftWheelMmps:  speedLeft,
			RightWheelMmps: speedRight,
		})
}

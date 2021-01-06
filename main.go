package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"os/signal"
	"strings"
	"syscall"
	"time"

	// qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type waHandler struct {
	c *whatsapp.Conn
}

type botChat struct {
	conn    *whatsapp.Conn
	message *whatsapp.TextMessage
}

type Api struct {
}

type botChatMessage struct {
	Time    string `bson:"time"`
	Id      string `bson:"id"`
	From    string `bson:"from"`
	Message string `bson:"message"`
}

var ctx = Context.Background()

func (b *botChat) connect() (*mongo.Database, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI("mongodb://192.168.32.102:27017")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return client.Database("whatsapp"), nil
}

func (api *Api) SayTime() string {
	var time1 = time.Now()
	if time1.Hour() > 18 {
		return "Malam"
	} else if time1.Hour() > 12 {
		return "Sore"
	} else if time1.Hour() > 10 {
		return "Siang"
	} else {
		return "Pagi"
	}
}

func (b *botChat) SayHi() {
	var api = Api{}
	//fmt.Printf("%v-%v-%v-%v-%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)
	if strings.Contains(b.message.Info.RemoteJid, "XXXX") {
		ContextInfo := whatsapp.ContextInfo{
			QuotedMessage:   nil,
			QuotedMessageID: "",
			Participant:     "", //Whot sent the original message
		}
		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: b.message.Info.RemoteJid,
			},
			ContextInfo: ContextInfo,
			Text:        "Hi...Selamat " + api.SayTime(),
		}
		msgId, err := b.conn.Send(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		} else {
			fmt.Println("Message Sent -> ID : " + msgId)
		}
	} else {
		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: b.message.Info.RemoteJid,
			},
			ContextInfo: b.message.ContextInfo,
			Text:        "Hi...Selamat " + api.SayTime(),
		}
		msgId, err := b.conn.Send(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		} else {
			fmt.Println("Message Sent -> ID : " + msgId)
		}
	}
}

func (b *botChat) HandleMessage(message whatsapp.TextMessage) {
	b.message = &message
	if strings.ToLower(message.Text) == "/hi" {
		b.SayHi()
	}
	//if message.Text == "/hi" || message.Info.FromMe == false {
	//	ContextInfo := whatsapp.ContextInfo{
	//		QuotedMessage:   nil,
	//		QuotedMessageID: "",
	//		Participant:     "", //Whot sent the original message
	//	}
	//
	//	msg := whatsapp.TextMessage{
	//		Info: whatsapp.MessageInfo{
	//			RemoteJid: message.Info.RemoteJid,
	//		},
	//		ContextInfo: ContextInfo,
	//		Text:        "Hi..Selamat Malam",
	//	}
	//
	//	msgId, err := b.conn.Send(msg)
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
	//		//os.Exit(1)
	//	} else {
	//		fmt.Println("Message Sent -> ID : " + msgId)
	//	}
	//}
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (wac *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	fmt.Printf("%v-%v-%v-%v-%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)
	var b = botChat{conn: wac.c}
	b.HandleMessage(message)
}

/*//Example for media handling. Video, Audio, Document are also possible in the same way
func (h *waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	data, err := message.Download()
	if err != nil {
		if err != whatsapp.ErrMediaDownloadFailedWith410 && err != whatsapp.ErrMediaDownloadFailedWith410 {
			return
		}
		if _, err = h.c.LoadMediaInfo(message.Info.SenderJid, message.Info.Id, strconv.FormatBool(message.Info.FromMe)); err == nil {
			data, err = message.Download()
			if err != nil {
				return
			}
		}
	}

	filename := fmt.Sprintf("%v/%v.%v", os.TempDir(), message.Info.Id, strings.Split(message.Type, "/")[1])
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return
	}
	_, err = file.Write(data)
	if err != nil {
		return
	}
	log.Printf("%v %v\n\timage reveived, saved at:%v\n", message.Info.Timestamp, message.Info.RemoteJid, filename)
}*/

func main() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			// terminal := qrcodeTerminal.New()
			// terminal.Get(<-qr).Print()
			qrToImg(<-qr)
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func qrToImg(qrCode string) {
	err := qrcode.WriteFile(qrCode, qrcode.Medium, 256, "qr.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo crear el archivo de imagen QR")
	}
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open("../whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create("../whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

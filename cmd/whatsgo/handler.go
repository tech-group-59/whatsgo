package main

import (
	"fmt"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"mime"
	"os"
	"strings"
)

func CreateHandler(fileFolder string, trackers []Tracker, config *Config) func(interface{}) {

	//var historySyncID int32
	//var startupTime = time.Now().Unix()

	handler := func(rawEvt interface{}) {
		switch evt := rawEvt.(type) {
		case *events.AppStateSyncComplete:
			if len(cli.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
				err := cli.SendPresence(types.PresenceAvailable)
				if err != nil {
					log.Warnf("Failed to send available presence: %v", err)
				} else {
					log.Infof("Marked self as available")
				}
			}
		case *events.Connected, *events.PushNameSetting:
			if len(cli.Store.PushName) == 0 {
				return
			}
			// Send presence available when connecting and when the pushname is changed.
			// This makes sure that outgoing messages always have the right pushname.
			err := cli.SendPresence(types.PresenceAvailable)
			if err != nil {
				log.Warnf("Failed to send available presence: %v", err)
			} else {
				log.Infof("Marked self as available")
			}
		case *events.StreamReplaced:
			os.Exit(0)
		case *events.Message:
			timestamp := evt.Info.Timestamp
			metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", timestamp)}
			if evt.Info.Type != "" {
				metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
			}
			if evt.Info.Category != "" {
				metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
			}
			if evt.IsViewOnce {
				metaParts = append(metaParts, "view once")
			}
			if evt.IsViewOnce {
				metaParts = append(metaParts, "ephemeral")
			}
			if evt.IsViewOnceV2 {
				metaParts = append(metaParts, "ephemeral (v2)")
			}
			if evt.IsDocumentWithCaption {
				metaParts = append(metaParts, "document with caption")
			}
			if evt.IsEdit {
				metaParts = append(metaParts, "edit")
			}

			log.Infof("Received message %s from %s (%s)", evt.Info.ID, evt.Info.SourceString(), strings.Join(metaParts, ", "))

			log.Infof(evt.Info.MessageSource.Chat.String())
			log.Infof(evt.Info.MessageSource.Sender.String())
			var sender = evt.Info.MessageSource.Sender.String()
			var chat = evt.Info.MessageSource.Chat.String()

			var trackable = config.IsChatTrackable(chat)

			var text string
			if trackable && evt.Info.Type == "text" {
				text = evt.Message.GetConversation()
				if text == "" && evt.Message.ExtendedTextMessage != nil {
					text = *evt.Message.ExtendedTextMessage.Text
				}
			}

			if evt.Message.GetPollUpdateMessage() != nil {
				decrypted, err := cli.DecryptPollVote(evt)
				if err != nil {
					log.Errorf("Failed to decrypt vote: %v", err)
				} else {
					log.Infof("Selected options in decrypted vote:")
					for _, option := range decrypted.SelectedOptions {
						log.Infof("- %X", option)
					}
				}
			} else if evt.Message.GetEncReactionMessage() != nil {
				decrypted, err := cli.DecryptReaction(evt)
				if err != nil {
					log.Errorf("Failed to decrypt encrypted reaction: %v", err)
				} else {
					log.Infof("Decrypted reaction: %+v", decrypted)
				}
			}

			date := timestamp.Format("02.01.2006")

			// Get the chat alias or ID
			folder := chat
			for _, c := range config.Chats {
				if c.ID == chat && c.Alias != "" {
					folder = c.Alias
					break
				}
			}

			metadata := MessageMetadata{
				Date:      date,
				Folder:    folder,
				Timestamp: timestamp,
			}

			var files []string
			img := evt.Message.GetImageMessage()
			if trackable && img != nil {
				data, err := cli.Download(img)
				if err != nil {
					log.Errorf("Failed to download image: %v", err)
					return
				}
				exts, _ := mime.ExtensionsByType(img.GetMimetype())
				ext := exts[0]
				if ext == ".jpe" {
					ext = ".jpg"
				}

				subFolder := fmt.Sprintf("%s/%s/%s", fileFolder, folder, date)
				path := fmt.Sprintf("%s/%s%s", subFolder, evt.Info.ID, ext)

				// Create sub folder if it doesn't exist
				if _, err := os.Stat(subFolder); os.IsNotExist(err) {
					err = os.MkdirAll(subFolder, 0755)
					if err != nil {
						log.Errorf("Failed to create subfolder: %v", err)
						return
					}
				}

				err = os.WriteFile(path, data, 0755)
				if err != nil {
					log.Errorf("Failed to save image: %v", err)
					return
				}
				text = img.GetCaption()
				files = append(files, path)

				if err != nil {
					log.Errorf("Failed to store file: %v", err)
				}

				log.Infof("Saved image in message to %s", path)
			}

			voice := evt.Message.GetAudioMessage()

			if trackable && voice != nil {
				data, err := cli.Download(voice)
				if err != nil {
					log.Errorf("Failed to download voice message: %v", err)
					return
				}

				subFolder := fmt.Sprintf("%s/%s/%s", fileFolder, folder, date)
				path := fmt.Sprintf("%s/%s.ogg", subFolder, evt.Info.ID)

				// Create sub folder if it doesn't exist
				if _, err := os.Stat(subFolder); os.IsNotExist(err) {
					err = os.MkdirAll(subFolder, 0755)
					if err != nil {
						log.Errorf("Failed to create subfolder: %v", err)
						return
					}
				}

				err = os.WriteFile(path, data, 0755)
				if err != nil {
					log.Errorf("Failed to save voice message: %v", err)
					return
				}
				files = append(files, path)

				log.Infof("Saved voice message in message to %s", path)
			}

			document := evt.Message.GetDocumentMessage()
			if trackable && document != nil {
				data, err := cli.Download(document)
				if err != nil {
					log.Errorf("Failed to download document: %v", err)
					return
				}

				exts, _ := mime.ExtensionsByType(document.GetMimetype())

				subFolder := fmt.Sprintf("%s/%s/%s", fileFolder, folder, date)
				path := fmt.Sprintf("%s/%s%s", subFolder, evt.Info.ID, exts[0])

				// Create sub folder if it doesn't exist
				if _, err := os.Stat(subFolder); os.IsNotExist(err) {
					err = os.MkdirAll(subFolder, 0755)
					if err != nil {
						log.Errorf("Failed to create subfolder: %v", err)
						return
					}
				}

				err = os.WriteFile(path, data, 0755)
				if err != nil {
					log.Errorf("Failed to save document: %v", err)
					return
				}
				files = append(files, path)

				log.Infof("Saved document in message to %s", path)
			}

			if trackable && (text != "" || len(files) > 0) {
				log.Infof("Tracking message from %s in chat %s", sender, chat)
				ProcessMessage(trackers, evt.Info.ID, sender, chat, text, timestamp.String(), files, metadata)
				log.Infof("Message text: %s", text)
			} else {
				log.Infof("Ignoring message from %s in chat %s", sender, chat)
			}

		case *events.Receipt:
			if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
				log.Infof("%v was read by %s at %s", evt.MessageIDs, evt.SourceString(), evt.Timestamp)
			} else if evt.Type == types.ReceiptTypeDelivered {
				log.Infof("%s was delivered to %s at %s", evt.MessageIDs[0], evt.SourceString(), evt.Timestamp)
			}
		case *events.Presence:
			if evt.Unavailable {
				if evt.LastSeen.IsZero() {
					log.Infof("%s is now offline", evt.From)
				} else {
					log.Infof("%s is now offline (last seen: %s)", evt.From, evt.LastSeen)
				}
			} else {
				log.Infof("%s is now online", evt.From)
			}
		case *events.HistorySync:
			log.Infof("Skip history sync event: %+v", evt)
		//id := atomic.AddInt32(&historySyncID, 1)
		//fileName := fmt.Sprintf("history-%d-%d.json", startupTime, id)
		//file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
		//if err != nil {
		//	log.Errorf("Failed to open file to write history sync: %v", err)
		//	return
		//}
		//enc := json.NewEncoder(file)
		//enc.SetIndent("", "  ")
		//err = enc.Encode(evt.Data)
		//if err != nil {
		//	log.Errorf("Failed to write history sync: %v", err)
		//	return
		//}
		//log.Infof("Wrote history sync to %s", fileName)
		//_ = file.Close()
		case *events.AppState:
			log.Debugf("App state event: %+v / %+v", evt.Index, evt.SyncActionValue)
		case *events.KeepAliveTimeout:
			log.Debugf("Keepalive timeout event: %+v", evt)
		case *events.KeepAliveRestored:
			log.Debugf("Keepalive restored")
		case *events.Blocklist:
			log.Infof("Blocklist event: %+v", evt)
		}
	}

	return handler
}

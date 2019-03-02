package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	// AnkiConnectAPI http api endpoint for connect to anki web.
	AnkiConnectAPI = "http://localhost:8765"

	defaultNoteTag = "from-yd" // from youdao translate
)

var ankiBackContentTmpl = `
<div align="left">
	<p>{{.Phonetic}}</p>
	<p>{{.Level}}</p>
	<p>{{.ECTrans}}</p>
</div>
<div align="left">
	<p><b>{{.CollinsTitle}}</b></p>
	{{range $entry := .CollinsEntries}}
		{{if $entry.HasSeeAlso}}
			<p>{{$entry.SeeAlso}}</p>
		{{else}}
			<p>
				{{$entry.Paraphrase}}<br>
				{{$entry.EngExampleSents}}<br>
				{{$entry.ChiExampleSents}}<br>
			</p>
    	{{end}}
	{{end}}
</div>
`

type (
	// AnkiNoteMeta meta info with anki web.
	AnkiNoteMeta struct {
		Deck  string
		Front string
		Back  string
		// TODO: option not work as expectation, Workaround:  use a temporary file: (~/tmp/yd.db that record the checksum for specific word.
		AllowDup bool
		Tags     []string
		AudioURL string
	}

	// AnkiAction anki action payload abstraction.
	AnkiAction struct {
		Action  string            `json:"action"`
		Version int               `json:"version"`
		Params  *AnkiActionParams `json:"params"`
	}

	// AnkiActionParams anki action params payload abstraction.
	AnkiActionParams struct {
		Note *AnkiActionNote `json:"note"`
	}

	// AnkiActionNote anki action note payload abstraction.
	AnkiActionNote struct {
		DeckName  string                 `json:"deckName"`
		ModelName string                 `json:"modelName"`
		Fields    *AnkiActionNoteFields  `json:"fields"`
		Options   *AnkiActionNoteOptions `json:"options"`
		Audio     *AnkiNoteAudio         `json:"audio"`
		Tags      []string               `json:"tags"`
	}

	//AnkiActionNoteFields anki action note fields payload abstraction.
	AnkiActionNoteFields struct {
		Front string `json:"Front"`
		Back  string `json:"Back"`
	}

	// AnkiActionNoteOptions anki action note options payload abstraction.
	AnkiActionNoteOptions struct {
		AllowDuplicate bool `json:"allowDuplicate "`
	}

	// AnkiNoteAudio anki action note audio payload abstraction.
	AnkiNoteAudio struct {
		URL      string   `json:"url"`
		Filename string   `json:"filename"`
		Fields   []string `json:"fields"`
		SkipHash string   `json:"skipHash"`
	}
)

// GetDeck get anki deck name from environment var.
func GetDeck() (string, error) {
	deck := os.Getenv("ANKI_DECK_NAME")
	if deck == "" {
		return "", fmt.Errorf("Anki Error: no deck name found, plz set env ANKI_DECK_NAME to your personal deck")
	}
	return deck, nil
}

func buildPayloadFromAnkiNote(note *AnkiNoteMeta) *AnkiAction {
	return &AnkiAction{
		Action:  "addNote",
		Version: 6,
		Params: &AnkiActionParams{
			Note: &AnkiActionNote{
				DeckName:  note.Deck,
				ModelName: "Basic",
				Fields: &AnkiActionNoteFields{
					Front: note.Front,
					Back:  note.Back,
				},
				Options: &AnkiActionNoteOptions{
					AllowDuplicate: note.AllowDup,
				},
				Audio: &AnkiNoteAudio{
					URL:      note.AudioURL,
					Filename: fmt.Sprintf("%s.mp3", note.Front),
					Fields:   []string{"Front"},
					SkipHash: "",
				},
				Tags: []string{defaultNoteTag},
			},
		},
	}
}

// AnkiClient http client that do short connection with anki api endpoint.
type AnkiClient struct {
	client   *http.Client
	endpoint string
}

// NewAnkiClient creates an new ank client.
func NewAnkiClient(endpoint string) *AnkiClient {
	return &AnkiClient{
		client:   &http.Client{Timeout: time.Second * 3},
		endpoint: endpoint,
	}
}

// AddNote add note to a specific deck.
func (a *AnkiClient) AddNote(note *AnkiNoteMeta) error {
	path := a.endpoint

	bs, err := json.Marshal(buildPayloadFromAnkiNote(note))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", path, bytes.NewReader(bs))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get an unexpected status code %d", resp.StatusCode)
	}
	return nil
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"
)

const (
	apiVendor     = "appstore"
	apiAppVersion = "2.4.0"
	apiClientFrom = "macdict"
	apiKeyFrom    = "mac.main"
	YouDaoAPI     = "http://dict.youdao.com"
)

// WordResp response abstraction with word translation from youdao api.
type (
	WordResp struct {
		word string
		EC   struct {
			ExamType []string `json:"exam_type"`
			Word     []struct {
				Trans []struct {
					Tr []struct {
						L struct {
							I []string `json:"i"`
						} `json:"l"`
					} `json:"tr"`
				} `json:"trs"`
				UKPhonetic string `json:"ukphone"`
				UKSpeech   string `json:"ukspeech"`
				USPhonetic string `json:"usphone"`
				USSpeech   string `json:"usspeech"`
			} `json:"word"`
		} `json:"ec"`
		Collins struct {
			CollinsEntries []struct {
				BasicEntries struct {
					BasicEntry []struct {
						CET      string `json:"cet"`
						HeadWord string `json:"headword"`
					} `json:"basic_entry"`
				} `json:"basic_entries"`
				Entries struct {
					Entry []struct {
						TranEntry []struct {
							ExampleSentences struct {
								Sentences []struct {
									ChineseSentence string `json:"chn_sent"`
									EnglishSentence string `json:"eng_sent"`
								} `json:"sent"`
							} `json:"exam_sents"`
							PosEntry struct {
								Pos     string `json:"pos"`
								PosTips string `json:"pos_tips"`
							} `json:"pos_entry"`
							Translation string `json:"tran"`
							SeeAlsos    struct {
								SeeAlso []struct {
									Seeword string `json:"seeword"`
								} `json:"seeAlso"`
								Seealso string `json:"seealso"`
							} `json:"seeAlsos"`
						} `json:"tran_entry"`
					} `json:"entry"`
				} `json:"entries"`
			} `json:"collins_entries"`
		} `json:"collins"`
	}

	// CollinsEntry collins translation entry
	CollinsEntry struct {
		Paraphrase      string
		EngExampleSents string
		ChiExampleSents string
		SeeAlso         string
		HasSeeAlso      bool
	}
)

// Word returns realistic word name.
func (w *WordResp) Word() string {
	return w.word
}

// HasECTrans indicates the word whether valid or not.
func (w *WordResp) HasECTrans() bool {
	return len(w.EC.Word) != 0
}

// HasLevel indicates the world has level info or not.
func (w *WordResp) HasLevel() bool {
	return len(w.EC.ExamType) != 0
}

// HasCollins indicates the word translation response has collins or not.
func (w *WordResp) HasCollins() bool {
	return len(w.Collins.CollinsEntries) != 0
}

// Phonetic returns phonetic symbol string.
func (w *WordResp) Phonetic() string {
	if !w.HasECTrans() {
		return ""
	}

	return fmt.Sprintf("英音： [%s] \t美音： [%s]", w.EC.Word[0].UKPhonetic, w.EC.Word[0].USPhonetic)
}

// Level returns level info.
func (w *WordResp) Level() string {
	if !w.HasLevel() {
		return ""
	}

	buf := new(strings.Builder)
	buf.WriteString("Level：")
	for _, et := range w.EC.ExamType {
		buf.WriteString(fmt.Sprintf("%s  ", et))
	}

	return buf.String()
}

// Invalid indicates the word whether invalid or not.
func (w *WordResp) Invalid() bool {
	return !w.HasECTrans() && !w.HasCollins()
}

// ECTrans english to chinese translation string representation.
func (w *WordResp) ECTrans() string {
	if !w.HasECTrans() {
		return ""
	}

	buf := new(strings.Builder)
	for _, tr := range w.EC.Word[0].Trans {
		buf.WriteString(fmt.Sprintf("%s\t", tr.Tr[0].L.I[0]))
	}

	return buf.String()
}

// USSpeechLink uk speech link.
func (w *WordResp) USSpeechLink() string {
	prefix := fmt.Sprintf("%s/dictvoice?audio=", YouDaoAPI)
	if !w.HasECTrans() || w.EC.Word[0].USSpeech == "" {
		return fmt.Sprintf("%s%s&type=2", prefix, w.Word())
	}

	return fmt.Sprintf("%s%s", prefix, w.EC.Word[0].USSpeech)
}

// UKSpeechLink us speech link.
func (w *WordResp) UKSpeechLink() string {
	prefix := fmt.Sprintf("%s/dictvoice?audio=", YouDaoAPI)
	if !w.HasECTrans() || w.EC.Word[0].UKSpeech == "" {
		return fmt.Sprintf("%s%s&type=2", prefix, w.Word())
	}

	return fmt.Sprintf("%s%s", prefix, w.EC.Word[0].UKSpeech)
}

// CollinsTitle returns collins title.
func (w *WordResp) CollinsTitle() string {
	if !w.HasCollins() {
		return ""
	}

	return fmt.Sprintf("柯林斯权威释义：\n\n")
}

// CollinsEntries returns collins translation entries.
func (w *WordResp) CollinsEntries() []*CollinsEntry {
	if !w.HasCollins() {
		return []*CollinsEntry{}
	}

	entries := make([]*CollinsEntry, 0)
	for i, te := range w.Collins.CollinsEntries[0].Entries.Entry {
		e := te.TranEntry[0]
		entry := &CollinsEntry{}
		if len(e.ExampleSentences.Sentences) > 0 {
			entry.Paraphrase = fmt.Sprintf("%d. %s %s %s\n", i+1, e.PosEntry.Pos, e.PosEntry.PosTips, e.Translation)
			entry.EngExampleSents = fmt.Sprintf("例：%s\n", e.ExampleSentences.Sentences[0].EnglishSentence)
			entry.ChiExampleSents = fmt.Sprintf("%s\n\n", e.ExampleSentences.Sentences[0].ChineseSentence)
		} else { // may be `see also` statements
			if len(e.SeeAlsos.SeeAlso) > 0 {
				entry.HasSeeAlso = true
				entry.SeeAlso = fmt.Sprintf("%d. See also：%s\n\n", i+1, e.SeeAlsos.SeeAlso[0].Seeword)
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

// CollinsTrans collins authority translation string representation.
func (w *WordResp) CollinsTrans() string {
	if !w.HasCollins() {
		return ""
	}

	buf := new(strings.Builder)
	buf.WriteString(w.CollinsTitle())
	for _, e := range w.CollinsEntries() {
		if e.HasSeeAlso {
			buf.WriteString(e.SeeAlso)
		} else {
			buf.WriteString(e.Paraphrase)
			buf.WriteString(e.EngExampleSents)
			buf.WriteString(e.ChiExampleSents)
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}

// AnkiBackContent generate the back content which used for anki note.
func (w *WordResp) AnkiBackContent() string {
	if w.Invalid() {
		return ""
	}

	buf := new(bytes.Buffer)
	tmpl, err := template.New("anki").Parse(ankiBackContentTmpl)
	if err != nil {
		return ""
	}

	tmp := struct {
		Phonetic, Level, ECTrans, CollinsTitle string
		CollinsEntries                         []*CollinsEntry
	}{
		Phonetic:       w.Phonetic(),
		Level:          w.Level(),
		ECTrans:        w.ECTrans(),
		CollinsEntries: w.CollinsEntries(),
		CollinsTitle:   w.CollinsTitle(),
	}
	if err = tmpl.Execute(buf, tmp); err != nil {
		fmt.Println(err)
		return ""
	}

	return buf.String()
}

// Output the output string where display in terminal.
func (w *WordResp) Output() string {
	/*
		word

		phonetic pronunciation level
		translation


		collins translation:
		1. ...

		2. ...

		3. ...
	*/
	if w.Invalid() {
		return fmt.Sprintf("may be invalid word")
	}

	buf := new(strings.Builder)
	// line includes word name and two new lines
	buf.WriteString(w.Word())
	buf.WriteByte('\n')
	buf.WriteByte('\n')

	// ec lines
	if w.HasECTrans() {
		buf.WriteString(w.Phonetic())
		buf.WriteByte('\t')
		buf.WriteString(w.Level())
		buf.WriteByte('\n')
		buf.WriteString(w.ECTrans())
	}

	if w.HasCollins() {
		buf.WriteString("\n\n")
		buf.WriteString(w.CollinsTrans())
	}

	return buf.String()
}

// String string representation.
func (w *WordResp) String() string {
	return w.Output()
}

// YouDaoAPIClient http client that do short connection with youdao api endpoint.
type YouDaoAPIClient struct {
	endpoint string
	client   *http.Client
}

// NewYouDaoAPIClient creates a new YouDaoAPIClient.
func NewYouDaoAPIClient(endpoint string) *YouDaoAPIClient {
	return &YouDaoAPIClient{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 3 * time.Second},
	}
}

// Translate do actual translating process.
func (yd *YouDaoAPIClient) Translate(word string) (*WordResp, error) {
	path := fmt.Sprintf("%s/jsonapi?q=%s&doctype=json&keyfrom=%s&vendor=%s&appVer=%s&client=%s&jsonversion=2",
		yd.endpoint,
		url.QueryEscape(word),
		apiKeyFrom,
		apiVendor,
		apiAppVersion,
		apiClientFrom,
	)

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := yd.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	wr := &WordResp{word: word}
	if err := json.Unmarshal(bs, wr); err != nil {
		return nil, err
	}

	if wr.Invalid() {
		return nil, fmt.Errorf("%s maybe a invalid word", wr.Word())
	}

	return wr, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiVendor     = "appstore"
	apiAppVersion = "2.4.0"
	apiClientFrom = "macdict"
	apiKeyFrom    = "mac.main"
	APIURL        = "http://dict.youdao.com"
)

// WordResp response abstraction with word translation from youdao api.
type WordResp struct {
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

// USSpeechParams uk speech link.
func (w *WordResp) USSpeechLink() string {
	prefix := fmt.Sprintf("%s/dictvoice?audio=", APIURL)
	if !w.HasECTrans() || w.EC.Word[0].USSpeech == "" {
		return fmt.Sprintf("%s%s&type=2", prefix, w.Word())
	}

	return fmt.Sprintf("%s%s", prefix, w.EC.Word[0].USSpeech)
}

// UKSpeechParams us speech link.
func (w *WordResp) UKSpeechLink() string {
	prefix := fmt.Sprintf("%s/dictvoice?audio=", APIURL)
	if !w.HasECTrans() || w.EC.Word[0].UKSpeech == "" {
		return fmt.Sprintf("%s%s&type=2", prefix, w.Word())
	}

	return fmt.Sprintf("%s%s", prefix, w.EC.Word[0].UKSpeech)
}

// CollinsTrans collins authority translation string representation.
func (w *WordResp) CollinsTrans() string {
	if !w.HasCollins() {
		return ""
	}

	buf := new(strings.Builder)
	buf.WriteString("柯林斯权威释义：\n\n")
	for i, te := range w.Collins.CollinsEntries[0].Entries.Entry {
		e := te.TranEntry[0]
		if len(e.ExampleSentences.Sentences) > 0 {
			buf.WriteString(fmt.Sprintf("%d. %s %s %s\n", i+1, e.PosEntry.Pos, e.PosEntry.PosTips, e.Translation))
			buf.WriteString(fmt.Sprintf("例：%s\n", e.ExampleSentences.Sentences[0].EnglishSentence))
			buf.WriteString(fmt.Sprintf("%s\n\n", e.ExampleSentences.Sentences[0].ChineseSentence))
		} else { // may be `see also` statements
			if len(e.SeeAlsos.SeeAlso) > 0 {
				buf.WriteString(fmt.Sprintf("%d See also：%s\n\n", i+1, e.SeeAlsos.SeeAlso[0].Seeword))
			}
		}
	}

	return buf.String()
}

func (w *WordResp) Output() string {
	/*
		word

		phonetic pronunciation level
		translation


		collins transtaion:
		1. ...

		2. ...

		3. ...
	*/
	if !w.HasECTrans() && !w.HasCollins() {
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
		client:   http.DefaultClient,
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

	return wr, nil
}

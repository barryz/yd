package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func init() {
	flag.Usage = usage
	flag.Parse()
}

const (
	apiVendor     = "appstore"
	apiAppVersion = "2.4.0"
	apiClientFrom = "macdict"
	apiKeyFrom    = "mac.main"
	APIURL        = "http://dict.youdao.com"
)

// WordResp response abstraction with word translation from youdao api.
type WordResp struct {
	EC struct {
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
			USPhonetic string `json:"usphone"`
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

// String for fmt package use directly.
func (w *WordResp) String() string {
	word := w.Collins.CollinsEntries[0].BasicEntries.BasicEntry[0].HeadWord
	buf := new(strings.Builder)
	buf.WriteString(word)
	buf.WriteByte('\n')
	buf.WriteByte('\n')

	buf.WriteString(fmt.Sprintf("英音： [%s] \t美音： [%s]\n", w.EC.Word[0].UKPhonetic, w.EC.Word[0].USPhonetic))

	for _, tr := range w.EC.Word[0].Trans {
		buf.WriteString(fmt.Sprintf("%s\t", tr.Tr[0].L.I[0]))
	}
	buf.WriteByte('\n')
	buf.WriteByte('\n')

	buf.WriteString("柯林斯权威释义：\n")

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
		word,
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

	wr := new(WordResp)
	if err := json.Unmarshal(bs, wr); err != nil {
		return nil, err
	}

	return wr, nil
}

var (
	word = flag.String("w", "", "word will translating")
)

func usage() {
	fmt.Fprintf(os.Stderr, "tr is translate command line program\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "tr [option]\n")
	flag.PrintDefaults()
	os.Exit(0)
}

func errOnExit(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func main() {
	if *word == "" {
		errOnExit(fmt.Errorf("you must speicify a word to translate"))
	}

	ydCli := NewYouDaoAPIClient(APIURL)
	resp, err := ydCli.Translate(*word)
	if err != nil {
		errOnExit(err)
	}

	fmt.Println(resp)
}

package bot

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/Avimitin/go-bot/internal/database"
)

func RegexKAR(msg *string, k *KnRType) (string, bool) {
	for key, rps := range *k {
		if strings.Contains(*msg, key) {
			if length := len(rps); length > 1 {
				rand.Seed(time.Now().Unix())
				return rps[rand.Intn(length)], true
			}
			return rps[0], true
		}
	}
	return "", false
}

// Function for set add new keyword and related reply.
func setKeywordIntoDB(keyword string, reply string, DB *sql.DB) (int, error) {
	if id, err := database.PeekKeywords(DB, keyword); id != -1 && err == nil {
		err = database.SetReply(DB, reply, id)
		if err != nil {
			log.Println("[ERR]Fail to set reply into database", err)
			return -1, err
		}
		return 0, nil
	} else if err != nil {
		log.Println("[ERR]Fail to get keyword in database", err)
		return -1, err
	}
	return database.AddKeywords(DB, keyword, reply)
}

func setKeywordIntoCFG(keyword string, reply string, k *KnRType) {
	if rps, exist := (*k)[keyword]; exist {
		rps = append(rps)
	} else {
		(*k)[keyword] = []string{reply}
	}
}

// Set can set given keyword and reply into memory and database
func Set(keyword string, reply string, c *Context) error {
	setKeywordIntoCFG(keyword, reply, c.KeywordReplies())
	id, err := setKeywordIntoDB(keyword, reply, c.DB())
	if id == -1 && err != nil {
		log.Println("[ERR]Fail to set keyword and reply", err)
		return err
	}
	return nil
}

type KeyWordNotFoundError struct{}

func (k *KeyWordNotFoundError) Error() string {
	return "Word not found."
}

func delKeywordAtDB(keyword string, db *sql.DB) error {
	ID, err := database.PeekKeywords(db, keyword)
	if err != nil {
		log.Println("[ERR]Fail to get keyword when delete keyword at db", err)
		return err
	}
	if ID == -1 {
		log.Println("[ERR]Keyword not found when delete keyword at db", err)
		return &KeyWordNotFoundError{}
	}
	err = database.DelReplyByKeyword(db, ID)
	if err != nil {
		log.Println("[ERR]Failed to delete reply when delete keyword at db", err)
		return err
	}
	return database.DelKeyword(db, ID)
}

func delKeywordAtCFG(keyword string, k *KnRType) error {
	_, e := (*k)[keyword]
	if e {
		delete(*k, keyword)
	}
	return &KeyWordNotFoundError{}
}

// Del will delete given keyword and it's associated replies
func Del(keyword string, c *Context) error {
	err := delKeywordAtCFG(keyword, c.KeywordReplies())
	if err != nil {
		log.Println("[ERR]Keyword not found")
		return err
	}
	err = delKeywordAtDB(keyword, c.DB())
	if err != nil {
		log.Println("[ERR]Fail to delKeywordAtDB", err)
		return err
	}
	return nil
}

// Load will load keyword and reply from database
func Load(db *sql.DB) (*KnRType, error) {
	kts, err := database.FetchKeyword(db)
	if err != nil {
		log.Println("[ERR]Error happen when Loading keyword", err)
		return nil, err
	}
	k := make(KnRType)
	for _, kt := range *kts {
		replies, err := database.GetReplyWithKey(db, kt.I)
		if err != nil {
			log.Println("[ERR]Failed to get reply when loading keyword", err)
			return nil, err
		}
		k[kt.K] = replies
	}
	return &k, nil
}

func ListKeywordAndReply(k *KnRType) string {
	text := "KEYS AND REPLIES:\n"
	for key, word := range *k {
		text += fmt.Sprintf("%s. K: %s\n", key, word)
		text += "\n"
	}
	return text
}

func DelReply(keyword string, reply string, c *Context) error {
	err := DelRepliesAtData(keyword, reply, c.KeywordReplies())
	if err != nil {
		return err
	}
	err = DelRepliesAtDatabase(reply, c.DB())
	if err != nil {
		return err
	}
	return nil
}

func DelRepliesAtData(keyword string, reply string, k *KnRType) error {
	replies, ok := (*k)[keyword]
	if !ok {
		return &KeyWordNotFoundError{}
	}
	for i, r := range replies {
		if r == reply {
			(*k)[keyword] = append(replies[:i], replies[i+1:]...)
			return nil
		}
	}
	return &KeyWordNotFoundError{}
}

func DelRepliesAtDatabase(reply string, db *sql.DB) error {
	i, err := database.PeekReply(db, reply)
	if err != nil {
		return err
	}
	err = database.DelReply(db, i)
	if err != nil {
		return err
	}
	return nil
}

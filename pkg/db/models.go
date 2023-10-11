package db

type Message struct {
	UserID     uint   `bson:"userid,omitempty"`
	ChatID     uint   `bson:"chatid,omitempty"`
	Text       string `bson:"text,omitempty"`
	BotMessage bool   `bson:"isbot,omitempty"`
}

type Translation struct {
	UserID     uint   `bson:"userid,omitempty"`
	ChatID     uint   `bson:"chatid,omitempty"`
	SourceText string `bson:"sourcetext,omitempty"`
	TargetText string `bson:"targettext,omitempty"`
	Source     string `bson:"source,omitempty"`
	Target     string `bson:"target,omitempty"`
}

type Config struct {
	UserID          uint   `bson:"userid,omitempty"`
	Source          string `bson:"source,omitempty"`
	Target          string `bson:"target,omitempty"`
	Mode            string `bson:"mode,omitempty"`
	TranslationWord string `bson:"translationWord,omitempty"`
}

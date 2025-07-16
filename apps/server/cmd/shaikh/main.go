package main

import (
	"fmt"
	"strings"

	"github.com/awbalessa/shaikh/apps/server/internal/stemmer"
)

func main() {
	text := `كان الطفل يركض في الحديقة بينما كانت الطيور تغرد فوق الأشجار. قرر الأب أن يأخذ أسرته في نزهة إلى البحر حيث يمكنهم الاستمتاع بنسيم الهواء العليل ومشاهدة الأمواج. قرأت الأم كتابًا عن التاريخ الإسلامي وجلست تتأمل في المعاني العميقة التي حملتها السطور. في المساء، تناولت العائلة العشاء معًا وتحدثوا عن أحلامهم المستقبلية وطموحاتهم. بالرغم من التعب، شعر الجميع بالسعادة والرضا عن يومهم الجميل والمليء بالحيوية.`

	stemmer := stemmer.NewArabicLightStemmer()
	final := []string{}

	for _, word := range strings.Fields(text) {
		stem := stemmer.LightStem(strings.TrimSpace(word))
		final = append(final, stem)
	}

	stemmedText := stemmer.LightStem("أحلامهم")
	fmt.Printf("Original: %s\n\nStemmed: %s\n\n", text, stemmedText)
}

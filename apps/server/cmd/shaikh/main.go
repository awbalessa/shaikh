package main

import (
	"fmt"
	"strings"

	"github.com/awbalessa/shaikh/apps/server/internal/stemmer"
)

func main() {
	diac_text := `كانَ الطِّفْلُ يَرْكُضُ في الحَدِيقَةِ بَيْنَما كانَتِ الطُّيُورُ تُغَرِّدُ فَوْقَ الأَشْجارِ. قَرَّرَ الأَبُ أَنْ يَأْخُذَ أُسْرَتَهُ في نُزْهَةٍ إِلَى البَحْرِ حَيْثُ يُمْكِنُهُمُ الاسْتِمْتاعُ بِنَسِيمِ الهَوَاءِ العَلِيلِ وَمُشَاهَدَةِ الأَمْواجِ. قَرَأَتِ الأُمُّ كِتابًا عَنِ التَّارِيخِ الإِسْلَامِيِّ وَجَلَسَتْ تَتَأَمَّلُ في المَعَانِي العَمِيقَةِ الَّتِي حَمَلَتْهَا السُّطُورُ. في المَسَاءِ، تَنَاوَلَتِ العَائِلَةُ العَشَاءَ مَعًا وَتَحَدَّثُوا عَنْ أَحْلَامِهِمُ المُسْتَقْبَلِيَّةِ وَطُمُوحَاتِهِمْ. بِالرَّغْمِ مِنَ التَّعَبِ، شَعَرَ الجَمِيعُ بِالسَّعَادَةِ وَالرِّضَا عَنْ يَوْمِهِمُ الجَمِيلِ وَالمَلِيءِ بِالحَيَوِيَّةِ.`

	stemmer := stemmer.NewArabicLightStemmer()

	finalWords := []string{}
	finalStems := []string{}
	for _, word := range strings.Fields(diac_text) {
		stripped := stemmer.WordProcessor.StripTashkeel(strings.TrimSpace(word))
		if stemmer.StopWordManager.IsStopword(stripped) {
			continue
		}
		finalWords = append(finalWords, stripped)
		finalStems = append(
			finalStems,
			stemmer.LightStem(stripped),
		)
	}

	fmt.Printf(
		"Words: %s\n\nStems: %s",
		strings.Join(finalWords, " "),
		strings.Join(finalStems, " "),
	)
}

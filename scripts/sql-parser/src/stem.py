from tashaphyne.stemming import ArabicLightStemmer

text = "كان الطفل يركض في الحديقة بينما كانت الطيور تغرد فوق الأشجار. قرر الأب أن يأخذ أسرته في نزهة إلى البحر حيث يمكنهم الاستمتاع بنسيم الهواء العليل ومشاهدة الأمواج. قرأت الأم كتابًا عن التاريخ الإسلامي وجلست تتأمل في المعاني العميقة التي حملتها السطور. في المساء، تناولت العائلة العشاء معًا وتحدثوا عن أحلامهم المستقبلية وطموحاتهم. بالرغم من التعب، شعر الجميع بالسعادة والرضا عن يومهم الجميل والمليء بالحيوية."

final = []
for word in text.split(" "):
    stemmed = ArabicLightStemmer().light_stem(
        word=word
    )
    final.append(stemmed)

print(f"Text: {text}")
print(f"Stemmed: {' '.join(final)}")

package domain

import (
	"database/sql/driver"
	"fmt"
)

type MessagesRole string

const (
	MessagesRoleUser     MessagesRole = "user"
	MessagesRoleModel    MessagesRole = "model"
	MessagesRoleFunction MessagesRole = "function"
)

func (e *MessagesRole) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = MessagesRole(s)
	case string:
		*e = MessagesRole(s)
	default:
		return fmt.Errorf("unsupported scan type for MessagesRole: %T", src)
	}
	return nil
}

type NullMessagesRole struct {
	MessagesRole MessagesRole
	Valid        bool // Valid is true if MessagesRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullMessagesRole) Scan(value any) error {
	if value == nil {
		ns.MessagesRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.MessagesRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullMessagesRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.MessagesRole), nil
}

type RagAyahNumber string

const (
	Ayah1   RagAyahNumber = "1"
	Ayah2   RagAyahNumber = "2"
	Ayah3   RagAyahNumber = "3"
	Ayah4   RagAyahNumber = "4"
	Ayah5   RagAyahNumber = "5"
	Ayah6   RagAyahNumber = "6"
	Ayah7   RagAyahNumber = "7"
	Ayah8   RagAyahNumber = "8"
	Ayah9   RagAyahNumber = "9"
	Ayah10  RagAyahNumber = "10"
	Ayah11  RagAyahNumber = "11"
	Ayah12  RagAyahNumber = "12"
	Ayah13  RagAyahNumber = "13"
	Ayah14  RagAyahNumber = "14"
	Ayah15  RagAyahNumber = "15"
	Ayah16  RagAyahNumber = "16"
	Ayah17  RagAyahNumber = "17"
	Ayah18  RagAyahNumber = "18"
	Ayah19  RagAyahNumber = "19"
	Ayah20  RagAyahNumber = "20"
	Ayah21  RagAyahNumber = "21"
	Ayah22  RagAyahNumber = "22"
	Ayah23  RagAyahNumber = "23"
	Ayah24  RagAyahNumber = "24"
	Ayah25  RagAyahNumber = "25"
	Ayah26  RagAyahNumber = "26"
	Ayah27  RagAyahNumber = "27"
	Ayah28  RagAyahNumber = "28"
	Ayah29  RagAyahNumber = "29"
	Ayah30  RagAyahNumber = "30"
	Ayah31  RagAyahNumber = "31"
	Ayah32  RagAyahNumber = "32"
	Ayah33  RagAyahNumber = "33"
	Ayah34  RagAyahNumber = "34"
	Ayah35  RagAyahNumber = "35"
	Ayah36  RagAyahNumber = "36"
	Ayah37  RagAyahNumber = "37"
	Ayah38  RagAyahNumber = "38"
	Ayah39  RagAyahNumber = "39"
	Ayah40  RagAyahNumber = "40"
	Ayah41  RagAyahNumber = "41"
	Ayah42  RagAyahNumber = "42"
	Ayah43  RagAyahNumber = "43"
	Ayah44  RagAyahNumber = "44"
	Ayah45  RagAyahNumber = "45"
	Ayah46  RagAyahNumber = "46"
	Ayah47  RagAyahNumber = "47"
	Ayah48  RagAyahNumber = "48"
	Ayah49  RagAyahNumber = "49"
	Ayah50  RagAyahNumber = "50"
	Ayah51  RagAyahNumber = "51"
	Ayah52  RagAyahNumber = "52"
	Ayah53  RagAyahNumber = "53"
	Ayah54  RagAyahNumber = "54"
	Ayah55  RagAyahNumber = "55"
	Ayah56  RagAyahNumber = "56"
	Ayah57  RagAyahNumber = "57"
	Ayah58  RagAyahNumber = "58"
	Ayah59  RagAyahNumber = "59"
	Ayah60  RagAyahNumber = "60"
	Ayah61  RagAyahNumber = "61"
	Ayah62  RagAyahNumber = "62"
	Ayah63  RagAyahNumber = "63"
	Ayah64  RagAyahNumber = "64"
	Ayah65  RagAyahNumber = "65"
	Ayah66  RagAyahNumber = "66"
	Ayah67  RagAyahNumber = "67"
	Ayah68  RagAyahNumber = "68"
	Ayah69  RagAyahNumber = "69"
	Ayah70  RagAyahNumber = "70"
	Ayah71  RagAyahNumber = "71"
	Ayah72  RagAyahNumber = "72"
	Ayah73  RagAyahNumber = "73"
	Ayah74  RagAyahNumber = "74"
	Ayah75  RagAyahNumber = "75"
	Ayah76  RagAyahNumber = "76"
	Ayah77  RagAyahNumber = "77"
	Ayah78  RagAyahNumber = "78"
	Ayah79  RagAyahNumber = "79"
	Ayah80  RagAyahNumber = "80"
	Ayah81  RagAyahNumber = "81"
	Ayah82  RagAyahNumber = "82"
	Ayah83  RagAyahNumber = "83"
	Ayah84  RagAyahNumber = "84"
	Ayah85  RagAyahNumber = "85"
	Ayah86  RagAyahNumber = "86"
	Ayah87  RagAyahNumber = "87"
	Ayah88  RagAyahNumber = "88"
	Ayah89  RagAyahNumber = "89"
	Ayah90  RagAyahNumber = "90"
	Ayah91  RagAyahNumber = "91"
	Ayah92  RagAyahNumber = "92"
	Ayah93  RagAyahNumber = "93"
	Ayah94  RagAyahNumber = "94"
	Ayah95  RagAyahNumber = "95"
	Ayah96  RagAyahNumber = "96"
	Ayah97  RagAyahNumber = "97"
	Ayah98  RagAyahNumber = "98"
	Ayah99  RagAyahNumber = "99"
	Ayah100 RagAyahNumber = "100"
	Ayah101 RagAyahNumber = "101"
	Ayah102 RagAyahNumber = "102"
	Ayah103 RagAyahNumber = "103"
	Ayah104 RagAyahNumber = "104"
	Ayah105 RagAyahNumber = "105"
	Ayah106 RagAyahNumber = "106"
	Ayah107 RagAyahNumber = "107"
	Ayah108 RagAyahNumber = "108"
	Ayah109 RagAyahNumber = "109"
	Ayah110 RagAyahNumber = "110"
	Ayah111 RagAyahNumber = "111"
	Ayah112 RagAyahNumber = "112"
	Ayah113 RagAyahNumber = "113"
	Ayah114 RagAyahNumber = "114"
	Ayah115 RagAyahNumber = "115"
	Ayah116 RagAyahNumber = "116"
	Ayah117 RagAyahNumber = "117"
	Ayah118 RagAyahNumber = "118"
	Ayah119 RagAyahNumber = "119"
	Ayah120 RagAyahNumber = "120"
	Ayah121 RagAyahNumber = "121"
	Ayah122 RagAyahNumber = "122"
	Ayah123 RagAyahNumber = "123"
	Ayah124 RagAyahNumber = "124"
	Ayah125 RagAyahNumber = "125"
	Ayah126 RagAyahNumber = "126"
	Ayah127 RagAyahNumber = "127"
	Ayah128 RagAyahNumber = "128"
	Ayah129 RagAyahNumber = "129"
	Ayah130 RagAyahNumber = "130"
	Ayah131 RagAyahNumber = "131"
	Ayah132 RagAyahNumber = "132"
	Ayah133 RagAyahNumber = "133"
	Ayah134 RagAyahNumber = "134"
	Ayah135 RagAyahNumber = "135"
	Ayah136 RagAyahNumber = "136"
	Ayah137 RagAyahNumber = "137"
	Ayah138 RagAyahNumber = "138"
	Ayah139 RagAyahNumber = "139"
	Ayah140 RagAyahNumber = "140"
	Ayah141 RagAyahNumber = "141"
	Ayah142 RagAyahNumber = "142"
	Ayah143 RagAyahNumber = "143"
	Ayah144 RagAyahNumber = "144"
	Ayah145 RagAyahNumber = "145"
	Ayah146 RagAyahNumber = "146"
	Ayah147 RagAyahNumber = "147"
	Ayah148 RagAyahNumber = "148"
	Ayah149 RagAyahNumber = "149"
	Ayah150 RagAyahNumber = "150"
	Ayah151 RagAyahNumber = "151"
	Ayah152 RagAyahNumber = "152"
	Ayah153 RagAyahNumber = "153"
	Ayah154 RagAyahNumber = "154"
	Ayah155 RagAyahNumber = "155"
	Ayah156 RagAyahNumber = "156"
	Ayah157 RagAyahNumber = "157"
	Ayah158 RagAyahNumber = "158"
	Ayah159 RagAyahNumber = "159"
	Ayah160 RagAyahNumber = "160"
	Ayah161 RagAyahNumber = "161"
	Ayah162 RagAyahNumber = "162"
	Ayah163 RagAyahNumber = "163"
	Ayah164 RagAyahNumber = "164"
	Ayah165 RagAyahNumber = "165"
	Ayah166 RagAyahNumber = "166"
	Ayah167 RagAyahNumber = "167"
	Ayah168 RagAyahNumber = "168"
	Ayah169 RagAyahNumber = "169"
	Ayah170 RagAyahNumber = "170"
	Ayah171 RagAyahNumber = "171"
	Ayah172 RagAyahNumber = "172"
	Ayah173 RagAyahNumber = "173"
	Ayah174 RagAyahNumber = "174"
	Ayah175 RagAyahNumber = "175"
	Ayah176 RagAyahNumber = "176"
	Ayah177 RagAyahNumber = "177"
	Ayah178 RagAyahNumber = "178"
	Ayah179 RagAyahNumber = "179"
	Ayah180 RagAyahNumber = "180"
	Ayah181 RagAyahNumber = "181"
	Ayah182 RagAyahNumber = "182"
	Ayah183 RagAyahNumber = "183"
	Ayah184 RagAyahNumber = "184"
	Ayah185 RagAyahNumber = "185"
	Ayah186 RagAyahNumber = "186"
	Ayah187 RagAyahNumber = "187"
	Ayah188 RagAyahNumber = "188"
	Ayah189 RagAyahNumber = "189"
	Ayah190 RagAyahNumber = "190"
	Ayah191 RagAyahNumber = "191"
	Ayah192 RagAyahNumber = "192"
	Ayah193 RagAyahNumber = "193"
	Ayah194 RagAyahNumber = "194"
	Ayah195 RagAyahNumber = "195"
	Ayah196 RagAyahNumber = "196"
	Ayah197 RagAyahNumber = "197"
	Ayah198 RagAyahNumber = "198"
	Ayah199 RagAyahNumber = "199"
	Ayah200 RagAyahNumber = "200"
	Ayah201 RagAyahNumber = "201"
	Ayah202 RagAyahNumber = "202"
	Ayah203 RagAyahNumber = "203"
	Ayah204 RagAyahNumber = "204"
	Ayah205 RagAyahNumber = "205"
	Ayah206 RagAyahNumber = "206"
	Ayah207 RagAyahNumber = "207"
	Ayah208 RagAyahNumber = "208"
	Ayah209 RagAyahNumber = "209"
	Ayah210 RagAyahNumber = "210"
	Ayah211 RagAyahNumber = "211"
	Ayah212 RagAyahNumber = "212"
	Ayah213 RagAyahNumber = "213"
	Ayah214 RagAyahNumber = "214"
	Ayah215 RagAyahNumber = "215"
	Ayah216 RagAyahNumber = "216"
	Ayah217 RagAyahNumber = "217"
	Ayah218 RagAyahNumber = "218"
	Ayah219 RagAyahNumber = "219"
	Ayah220 RagAyahNumber = "220"
	Ayah221 RagAyahNumber = "221"
	Ayah222 RagAyahNumber = "222"
	Ayah223 RagAyahNumber = "223"
	Ayah224 RagAyahNumber = "224"
	Ayah225 RagAyahNumber = "225"
	Ayah226 RagAyahNumber = "226"
	Ayah227 RagAyahNumber = "227"
	Ayah228 RagAyahNumber = "228"
	Ayah229 RagAyahNumber = "229"
	Ayah230 RagAyahNumber = "230"
	Ayah231 RagAyahNumber = "231"
	Ayah232 RagAyahNumber = "232"
	Ayah233 RagAyahNumber = "233"
	Ayah234 RagAyahNumber = "234"
	Ayah235 RagAyahNumber = "235"
	Ayah236 RagAyahNumber = "236"
	Ayah237 RagAyahNumber = "237"
	Ayah238 RagAyahNumber = "238"
	Ayah239 RagAyahNumber = "239"
	Ayah240 RagAyahNumber = "240"
	Ayah241 RagAyahNumber = "241"
	Ayah242 RagAyahNumber = "242"
	Ayah243 RagAyahNumber = "243"
	Ayah244 RagAyahNumber = "244"
	Ayah245 RagAyahNumber = "245"
	Ayah246 RagAyahNumber = "246"
	Ayah247 RagAyahNumber = "247"
	Ayah248 RagAyahNumber = "248"
	Ayah249 RagAyahNumber = "249"
	Ayah250 RagAyahNumber = "250"
	Ayah251 RagAyahNumber = "251"
	Ayah252 RagAyahNumber = "252"
	Ayah253 RagAyahNumber = "253"
	Ayah254 RagAyahNumber = "254"
	Ayah255 RagAyahNumber = "255"
	Ayah256 RagAyahNumber = "256"
	Ayah257 RagAyahNumber = "257"
	Ayah258 RagAyahNumber = "258"
	Ayah259 RagAyahNumber = "259"
	Ayah260 RagAyahNumber = "260"
	Ayah261 RagAyahNumber = "261"
	Ayah262 RagAyahNumber = "262"
	Ayah263 RagAyahNumber = "263"
	Ayah264 RagAyahNumber = "264"
	Ayah265 RagAyahNumber = "265"
	Ayah266 RagAyahNumber = "266"
	Ayah267 RagAyahNumber = "267"
	Ayah268 RagAyahNumber = "268"
	Ayah269 RagAyahNumber = "269"
	Ayah270 RagAyahNumber = "270"
	Ayah271 RagAyahNumber = "271"
	Ayah272 RagAyahNumber = "272"
	Ayah273 RagAyahNumber = "273"
	Ayah274 RagAyahNumber = "274"
	Ayah275 RagAyahNumber = "275"
	Ayah276 RagAyahNumber = "276"
	Ayah277 RagAyahNumber = "277"
	Ayah278 RagAyahNumber = "278"
	Ayah279 RagAyahNumber = "279"
	Ayah280 RagAyahNumber = "280"
	Ayah281 RagAyahNumber = "281"
	Ayah282 RagAyahNumber = "282"
	Ayah283 RagAyahNumber = "283"
	Ayah284 RagAyahNumber = "284"
	Ayah285 RagAyahNumber = "285"
	Ayah286 RagAyahNumber = "286"
)

func (e *RagAyahNumber) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = RagAyahNumber(s)
	case string:
		*e = RagAyahNumber(s)
	default:
		return fmt.Errorf("unsupported scan type for RagAyahNumber: %T", src)
	}
	return nil
}

type NullRagAyah struct {
	Ayah  RagAyahNumber
	Valid bool // Valid is true if RagAyahNumber is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRagAyah) Scan(value any) error {
	if value == nil {
		ns.Ayah, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Ayah.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRagAyah) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Ayah), nil
}

type RagContentType string

const (
	ContentTypeTafsir RagContentType = "tafsir"
)

func (e *RagContentType) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = RagContentType(s)
	case string:
		*e = RagContentType(s)
	default:
		return fmt.Errorf("unsupported scan type for RagContentType: %T", src)
	}
	return nil
}

type NullRagContentType struct {
	ContentType RagContentType
	Valid       bool // Valid is true if RagContentType is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRagContentType) Scan(value any) error {
	if value == nil {
		ns.ContentType, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.ContentType.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRagContentType) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.ContentType), nil
}

type RagGranularity string

const (
	GranularityPhrase RagGranularity = "phrase"
	GranularityAyah   RagGranularity = "ayah"
	GranularitySurah  RagGranularity = "surah"
	GranularityQuran  RagGranularity = "quran"
)

func (e *RagGranularity) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = RagGranularity(s)
	case string:
		*e = RagGranularity(s)
	default:
		return fmt.Errorf("unsupported scan type for RagGranularity: %T", src)
	}
	return nil
}

type NullRagGranularity struct {
	RagGranularity RagGranularity
	Valid          bool // Valid is true if RagGranularity is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRagGranularity) Scan(value any) error {
	if value == nil {
		ns.RagGranularity, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.RagGranularity.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRagGranularity) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.RagGranularity), nil
}

type RagSource string

const (
	SourceTafsirIbnKathir RagSource = "Tafsir Ibn Kathir"
)

func (e *RagSource) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = RagSource(s)
	case string:
		*e = RagSource(s)
	default:
		return fmt.Errorf("unsupported scan type for RagSource: %T", src)
	}
	return nil
}

type NullRagSource struct {
	Source RagSource
	Valid  bool // Valid is true if RagSource is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRagSource) Scan(value any) error {
	if value == nil {
		ns.Source, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.Source.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRagSource) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.Source), nil
}

type RagSurahNumber string

const (
	Surah1   RagSurahNumber = "1"
	Surah2   RagSurahNumber = "2"
	Surah3   RagSurahNumber = "3"
	Surah4   RagSurahNumber = "4"
	Surah5   RagSurahNumber = "5"
	Surah6   RagSurahNumber = "6"
	Surah7   RagSurahNumber = "7"
	Surah8   RagSurahNumber = "8"
	Surah9   RagSurahNumber = "9"
	Surah10  RagSurahNumber = "10"
	Surah11  RagSurahNumber = "11"
	Surah12  RagSurahNumber = "12"
	Surah13  RagSurahNumber = "13"
	Surah14  RagSurahNumber = "14"
	Surah15  RagSurahNumber = "15"
	Surah16  RagSurahNumber = "16"
	Surah17  RagSurahNumber = "17"
	Surah18  RagSurahNumber = "18"
	Surah19  RagSurahNumber = "19"
	Surah20  RagSurahNumber = "20"
	Surah21  RagSurahNumber = "21"
	Surah22  RagSurahNumber = "22"
	Surah23  RagSurahNumber = "23"
	Surah24  RagSurahNumber = "24"
	Surah25  RagSurahNumber = "25"
	Surah26  RagSurahNumber = "26"
	Surah27  RagSurahNumber = "27"
	Surah28  RagSurahNumber = "28"
	Surah29  RagSurahNumber = "29"
	Surah30  RagSurahNumber = "30"
	Surah31  RagSurahNumber = "31"
	Surah32  RagSurahNumber = "32"
	Surah33  RagSurahNumber = "33"
	Surah34  RagSurahNumber = "34"
	Surah35  RagSurahNumber = "35"
	Surah36  RagSurahNumber = "36"
	Surah37  RagSurahNumber = "37"
	Surah38  RagSurahNumber = "38"
	Surah39  RagSurahNumber = "39"
	Surah40  RagSurahNumber = "40"
	Surah41  RagSurahNumber = "41"
	Surah42  RagSurahNumber = "42"
	Surah43  RagSurahNumber = "43"
	Surah44  RagSurahNumber = "44"
	Surah45  RagSurahNumber = "45"
	Surah46  RagSurahNumber = "46"
	Surah47  RagSurahNumber = "47"
	Surah48  RagSurahNumber = "48"
	Surah49  RagSurahNumber = "49"
	Surah50  RagSurahNumber = "50"
	Surah51  RagSurahNumber = "51"
	Surah52  RagSurahNumber = "52"
	Surah53  RagSurahNumber = "53"
	Surah54  RagSurahNumber = "54"
	Surah55  RagSurahNumber = "55"
	Surah56  RagSurahNumber = "56"
	Surah57  RagSurahNumber = "57"
	Surah58  RagSurahNumber = "58"
	Surah59  RagSurahNumber = "59"
	Surah60  RagSurahNumber = "60"
	Surah61  RagSurahNumber = "61"
	Surah62  RagSurahNumber = "62"
	Surah63  RagSurahNumber = "63"
	Surah64  RagSurahNumber = "64"
	Surah65  RagSurahNumber = "65"
	Surah66  RagSurahNumber = "66"
	Surah67  RagSurahNumber = "67"
	Surah68  RagSurahNumber = "68"
	Surah69  RagSurahNumber = "69"
	Surah70  RagSurahNumber = "70"
	Surah71  RagSurahNumber = "71"
	Surah72  RagSurahNumber = "72"
	Surah73  RagSurahNumber = "73"
	Surah74  RagSurahNumber = "74"
	Surah75  RagSurahNumber = "75"
	Surah76  RagSurahNumber = "76"
	Surah77  RagSurahNumber = "77"
	Surah78  RagSurahNumber = "78"
	Surah79  RagSurahNumber = "79"
	Surah80  RagSurahNumber = "80"
	Surah81  RagSurahNumber = "81"
	Surah82  RagSurahNumber = "82"
	Surah83  RagSurahNumber = "83"
	Surah84  RagSurahNumber = "84"
	Surah85  RagSurahNumber = "85"
	Surah86  RagSurahNumber = "86"
	Surah87  RagSurahNumber = "87"
	Surah88  RagSurahNumber = "88"
	Surah89  RagSurahNumber = "89"
	Surah90  RagSurahNumber = "90"
	Surah91  RagSurahNumber = "91"
	Surah92  RagSurahNumber = "92"
	Surah93  RagSurahNumber = "93"
	Surah94  RagSurahNumber = "94"
	Surah95  RagSurahNumber = "95"
	Surah96  RagSurahNumber = "96"
	Surah97  RagSurahNumber = "97"
	Surah98  RagSurahNumber = "98"
	Surah99  RagSurahNumber = "99"
	Surah100 RagSurahNumber = "100"
	Surah101 RagSurahNumber = "101"
	Surah102 RagSurahNumber = "102"
	Surah103 RagSurahNumber = "103"
	Surah104 RagSurahNumber = "104"
	Surah105 RagSurahNumber = "105"
	Surah106 RagSurahNumber = "106"
	Surah107 RagSurahNumber = "107"
	Surah108 RagSurahNumber = "108"
	Surah109 RagSurahNumber = "109"
	Surah110 RagSurahNumber = "110"
	Surah111 RagSurahNumber = "111"
	Surah112 RagSurahNumber = "112"
	Surah113 RagSurahNumber = "113"
	Surah114 RagSurahNumber = "114"
)

func (e *RagSurahNumber) Scan(src any) error {
	switch s := src.(type) {
	case []byte:
		*e = RagSurahNumber(s)
	case string:
		*e = RagSurahNumber(s)
	default:
		return fmt.Errorf("unsupported scan type for RagSurahNumber: %T", src)
	}
	return nil
}

type NullRagSurah struct {
	SurahNumber RagSurahNumber
	Valid       bool // Valid is true if RagSurahNumber is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullRagSurah) Scan(value any) error {
	if value == nil {
		ns.SurahNumber, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.SurahNumber.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullRagSurah) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.SurahNumber), nil
}

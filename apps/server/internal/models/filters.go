package models

import "github.com/awbalessa/shaikh/apps/server/internal/database"

const (
	LabelContentTypeTafsir                LabelContentType = 1
	LabelSourceTafsirIbnKathir            LabelSource      = 101
	LabelSourceTafsirAlTabari             LabelSource      = 102
	LabelSourceTafsirAlQurtubi            LabelSource      = 103
	LabelSourceTafsirAlBaghawi            LabelSource      = 104
	LabelSourceTafsirAlSaadi              LabelSource      = 105
	LabelSourceTafsirAlMuyassar           LabelSource      = 106
	LabelSourceTafsirAlWasit              LabelSource      = 107
	LabelSourceTafsirAlJalalayn           LabelSource      = 108
	LabelSurahNumberOne                   LabelSurahNumber = 1001
	LabelSurahNumberTwo                   LabelSurahNumber = 1002
	LabelSurahNumberThree                 LabelSurahNumber = 1003
	LabelSurahNumberFour                  LabelSurahNumber = 1004
	LabelSurahNumberFive                  LabelSurahNumber = 1005
	LabelSurahNumberSix                   LabelSurahNumber = 1006
	LabelSurahNumberSeven                 LabelSurahNumber = 1007
	LabelSurahNumberEight                 LabelSurahNumber = 1008
	LabelSurahNumberNine                  LabelSurahNumber = 1009
	LabelSurahNumberTen                   LabelSurahNumber = 1010
	LabelSurahNumberEleven                LabelSurahNumber = 1011
	LabelSurahNumberTwelve                LabelSurahNumber = 1012
	LabelSurahNumberThirteen              LabelSurahNumber = 1013
	LabelSurahNumberFourteen              LabelSurahNumber = 1014
	LabelSurahNumberFifteen               LabelSurahNumber = 1015
	LabelSurahNumberSixteen               LabelSurahNumber = 1016
	LabelSurahNumberSeventeen             LabelSurahNumber = 1017
	LabelSurahNumberEighteen              LabelSurahNumber = 1018
	LabelSurahNumberNineteen              LabelSurahNumber = 1019
	LabelSurahNumberTwenty                LabelSurahNumber = 1020
	LabelSurahNumberTwentyOne             LabelSurahNumber = 1021
	LabelSurahNumberTwentyTwo             LabelSurahNumber = 1022
	LabelSurahNumberTwentyThree           LabelSurahNumber = 1023
	LabelSurahNumberTwentyFour            LabelSurahNumber = 1024
	LabelSurahNumberTwentyFive            LabelSurahNumber = 1025
	LabelSurahNumberTwentySix             LabelSurahNumber = 1026
	LabelSurahNumberTwentySeven           LabelSurahNumber = 1027
	LabelSurahNumberTwentyEight           LabelSurahNumber = 1028
	LabelSurahNumberTwentyNine            LabelSurahNumber = 1029
	LabelSurahNumberThirty                LabelSurahNumber = 1030
	LabelSurahNumberThirtyOne             LabelSurahNumber = 1031
	LabelSurahNumberThirtyTwo             LabelSurahNumber = 1032
	LabelSurahNumberThirtyThree           LabelSurahNumber = 1033
	LabelSurahNumberThirtyFour            LabelSurahNumber = 1034
	LabelSurahNumberThirtyFive            LabelSurahNumber = 1035
	LabelSurahNumberThirtySix             LabelSurahNumber = 1036
	LabelSurahNumberThirtySeven           LabelSurahNumber = 1037
	LabelSurahNumberThirtyEight           LabelSurahNumber = 1038
	LabelSurahNumberThirtyNine            LabelSurahNumber = 1039
	LabelSurahNumberFourty                LabelSurahNumber = 1040
	LabelSurahNumberFourtyOne             LabelSurahNumber = 1041
	LabelSurahNumberFourtyTwo             LabelSurahNumber = 1042
	LabelSurahNumberFourtyThree           LabelSurahNumber = 1043
	LabelSurahNumberFourtyFour            LabelSurahNumber = 1044
	LabelSurahNumberFourtyFive            LabelSurahNumber = 1045
	LabelSurahNumberFourtySix             LabelSurahNumber = 1046
	LabelSurahNumberFourtySeven           LabelSurahNumber = 1047
	LabelSurahNumberFourtyEight           LabelSurahNumber = 1048
	LabelSurahNumberFourtyNine            LabelSurahNumber = 1049
	LabelSurahNumberFifty                 LabelSurahNumber = 1050
	LabelSurahNumberFiftyOne              LabelSurahNumber = 1051
	LabelSurahNumberFiftyTwo              LabelSurahNumber = 1052
	LabelSurahNumberFiftyThree            LabelSurahNumber = 1053
	LabelSurahNumberFiftyFour             LabelSurahNumber = 1054
	LabelSurahNumberFiftyFive             LabelSurahNumber = 1055
	LabelSurahNumberFiftySix              LabelSurahNumber = 1056
	LabelSurahNumberFiftySeven            LabelSurahNumber = 1057
	LabelSurahNumberFiftyEight            LabelSurahNumber = 1058
	LabelSurahNumberFiftyNine             LabelSurahNumber = 1059
	LabelSurahNumberSixty                 LabelSurahNumber = 1060
	LabelSurahNumberSixtyOne              LabelSurahNumber = 1061
	LabelSurahNumberSixtyTwo              LabelSurahNumber = 1062
	LabelSurahNumberSixtyThree            LabelSurahNumber = 1063
	LabelSurahNumberSixtyFour             LabelSurahNumber = 1064
	LabelSurahNumberSixtyFive             LabelSurahNumber = 1065
	LabelSurahNumberSixtySix              LabelSurahNumber = 1066
	LabelSurahNumberSixtySeven            LabelSurahNumber = 1067
	LabelSurahNumberSixtyEight            LabelSurahNumber = 1068
	LabelSurahNumberSixtyNine             LabelSurahNumber = 1069
	LabelSurahNumberSeventy               LabelSurahNumber = 1070
	LabelSurahNumberSeventyOne            LabelSurahNumber = 1071
	LabelSurahNumberSeventyTwo            LabelSurahNumber = 1072
	LabelSurahNumberSeventyThree          LabelSurahNumber = 1073
	LabelSurahNumberSeventyFour           LabelSurahNumber = 1074
	LabelSurahNumberSeventyFive           LabelSurahNumber = 1075
	LabelSurahNumberSeventySix            LabelSurahNumber = 1076
	LabelSurahNumberSeventySeven          LabelSurahNumber = 1077
	LabelSurahNumberSeventyEight          LabelSurahNumber = 1078
	LabelSurahNumberSeventyNine           LabelSurahNumber = 1079
	LabelSurahNumberEighty                LabelSurahNumber = 1080
	LabelSurahNumberEightyOne             LabelSurahNumber = 1081
	LabelSurahNumberEightyTwo             LabelSurahNumber = 1082
	LabelSurahNumberEightyThree           LabelSurahNumber = 1083
	LabelSurahNumberEightyFour            LabelSurahNumber = 1084
	LabelSurahNumberEightyFive            LabelSurahNumber = 1085
	LabelSurahNumberEightySix             LabelSurahNumber = 1086
	LabelSurahNumberEightySeven           LabelSurahNumber = 1087
	LabelSurahNumberEightyEight           LabelSurahNumber = 1088
	LabelSurahNumberEightyNine            LabelSurahNumber = 1089
	LabelSurahNumberNinety                LabelSurahNumber = 1090
	LabelSurahNumberNinetyOne             LabelSurahNumber = 1091
	LabelSurahNumberNinetyTwo             LabelSurahNumber = 1092
	LabelSurahNumberNinetyThree           LabelSurahNumber = 1093
	LabelSurahNumberNinetyFour            LabelSurahNumber = 1094
	LabelSurahNumberNinetyFive            LabelSurahNumber = 1095
	LabelSurahNumberNinetySix             LabelSurahNumber = 1096
	LabelSurahNumberNinetySeven           LabelSurahNumber = 1097
	LabelSurahNumberNinetyEight           LabelSurahNumber = 1098
	LabelSurahNumberNinetyNine            LabelSurahNumber = 1099
	LabelSurahNumberOneHundred            LabelSurahNumber = 1100
	LabelSurahNumberOneHundredOne         LabelSurahNumber = 1101
	LabelSurahNumberOneHundredTwo         LabelSurahNumber = 1102
	LabelSurahNumberOneHundredThree       LabelSurahNumber = 1103
	LabelSurahNumberOneHundredFour        LabelSurahNumber = 1104
	LabelSurahNumberOneHundredFive        LabelSurahNumber = 1105
	LabelSurahNumberOneHundredSix         LabelSurahNumber = 1106
	LabelSurahNumberOneHundredSeven       LabelSurahNumber = 1107
	LabelSurahNumberOneHundredEight       LabelSurahNumber = 1108
	LabelSurahNumberOneHundredNine        LabelSurahNumber = 1109
	LabelSurahNumberOneHundredTen         LabelSurahNumber = 1110
	LabelSurahNumberOneHundredEleven      LabelSurahNumber = 1111
	LabelSurahNumberOneHundredTwelve      LabelSurahNumber = 1112
	LabelSurahNumberOneHundredThirteen    LabelSurahNumber = 1113
	LabelSurahNumberOneHundredFourteen    LabelSurahNumber = 1114
	LabelAyahNumberOne                    LabelAyahNumber  = 2001
	LabelAyahNumberTwo                    LabelAyahNumber  = 2002
	LabelAyahNumberThree                  LabelAyahNumber  = 2003
	LabelAyahNumberFour                   LabelAyahNumber  = 2004
	LabelAyahNumberFive                   LabelAyahNumber  = 2005
	LabelAyahNumberSix                    LabelAyahNumber  = 2006
	LabelAyahNumberSeven                  LabelAyahNumber  = 2007
	LabelAyahNumberEight                  LabelAyahNumber  = 2008
	LabelAyahNumberNine                   LabelAyahNumber  = 2009
	LabelAyahNumberTen                    LabelAyahNumber  = 2010
	LabelAyahNumberEleven                 LabelAyahNumber  = 2011
	LabelAyahNumberTwelve                 LabelAyahNumber  = 2012
	LabelAyahNumberThirteen               LabelAyahNumber  = 2013
	LabelAyahNumberFourteen               LabelAyahNumber  = 2014
	LabelAyahNumberFifteen                LabelAyahNumber  = 2015
	LabelAyahNumberSixteen                LabelAyahNumber  = 2016
	LabelAyahNumberSeventeen              LabelAyahNumber  = 2017
	LabelAyahNumberEighteen               LabelAyahNumber  = 2018
	LabelAyahNumberNineteen               LabelAyahNumber  = 2019
	LabelAyahNumberTwenty                 LabelAyahNumber  = 2020
	LabelAyahNumberTwentyOne              LabelAyahNumber  = 2021
	LabelAyahNumberTwentyTwo              LabelAyahNumber  = 2022
	LabelAyahNumberTwentyThree            LabelAyahNumber  = 2023
	LabelAyahNumberTwentyFour             LabelAyahNumber  = 2024
	LabelAyahNumberTwentyFive             LabelAyahNumber  = 2025
	LabelAyahNumberTwentySix              LabelAyahNumber  = 2026
	LabelAyahNumberTwentySeven            LabelAyahNumber  = 2027
	LabelAyahNumberTwentyEight            LabelAyahNumber  = 2028
	LabelAyahNumberTwentyNine             LabelAyahNumber  = 2029
	LabelAyahNumberThirty                 LabelAyahNumber  = 2030
	LabelAyahNumberThirtyOne              LabelAyahNumber  = 2031
	LabelAyahNumberThirtyTwo              LabelAyahNumber  = 2032
	LabelAyahNumberThirtyThree            LabelAyahNumber  = 2033
	LabelAyahNumberThirtyFour             LabelAyahNumber  = 2034
	LabelAyahNumberThirtyFive             LabelAyahNumber  = 2035
	LabelAyahNumberThirtySix              LabelAyahNumber  = 2036
	LabelAyahNumberThirtySeven            LabelAyahNumber  = 2037
	LabelAyahNumberThirtyEight            LabelAyahNumber  = 2038
	LabelAyahNumberThirtyNine             LabelAyahNumber  = 2039
	LabelAyahNumberFourty                 LabelAyahNumber  = 2040
	LabelAyahNumberFourtyOne              LabelAyahNumber  = 2041
	LabelAyahNumberFourtyTwo              LabelAyahNumber  = 2042
	LabelAyahNumberFourtyThree            LabelAyahNumber  = 2043
	LabelAyahNumberFourtyFour             LabelAyahNumber  = 2044
	LabelAyahNumberFourtyFive             LabelAyahNumber  = 2045
	LabelAyahNumberFourtySix              LabelAyahNumber  = 2046
	LabelAyahNumberFourtySeven            LabelAyahNumber  = 2047
	LabelAyahNumberFourtyEight            LabelAyahNumber  = 2048
	LabelAyahNumberFourtyNine             LabelAyahNumber  = 2049
	LabelAyahNumberFifty                  LabelAyahNumber  = 2050
	LabelAyahNumberFiftyOne               LabelAyahNumber  = 2051
	LabelAyahNumberFiftyTwo               LabelAyahNumber  = 2052
	LabelAyahNumberFiftyThree             LabelAyahNumber  = 2053
	LabelAyahNumberFiftyFour              LabelAyahNumber  = 2054
	LabelAyahNumberFiftyFive              LabelAyahNumber  = 2055
	LabelAyahNumberFiftySix               LabelAyahNumber  = 2056
	LabelAyahNumberFiftySeven             LabelAyahNumber  = 2057
	LabelAyahNumberFiftyEight             LabelAyahNumber  = 2058
	LabelAyahNumberFiftyNine              LabelAyahNumber  = 2059
	LabelAyahNumberSixty                  LabelAyahNumber  = 2060
	LabelAyahNumberSixtyOne               LabelAyahNumber  = 2061
	LabelAyahNumberSixtyTwo               LabelAyahNumber  = 2062
	LabelAyahNumberSixtyThree             LabelAyahNumber  = 2063
	LabelAyahNumberSixtyFour              LabelAyahNumber  = 2064
	LabelAyahNumberSixtyFive              LabelAyahNumber  = 2065
	LabelAyahNumberSixtySix               LabelAyahNumber  = 2066
	LabelAyahNumberSixtySeven             LabelAyahNumber  = 2067
	LabelAyahNumberSixtyEight             LabelAyahNumber  = 2068
	LabelAyahNumberSixtyNine              LabelAyahNumber  = 2069
	LabelAyahNumberSeventy                LabelAyahNumber  = 2070
	LabelAyahNumberSeventyOne             LabelAyahNumber  = 2071
	LabelAyahNumberSeventyTwo             LabelAyahNumber  = 2072
	LabelAyahNumberSeventyThree           LabelAyahNumber  = 2073
	LabelAyahNumberSeventyFour            LabelAyahNumber  = 2074
	LabelAyahNumberSeventyFive            LabelAyahNumber  = 2075
	LabelAyahNumberSeventySix             LabelAyahNumber  = 2076
	LabelAyahNumberSeventySeven           LabelAyahNumber  = 2077
	LabelAyahNumberSeventyEight           LabelAyahNumber  = 2078
	LabelAyahNumberSeventyNine            LabelAyahNumber  = 2079
	LabelAyahNumberEighty                 LabelAyahNumber  = 2080
	LabelAyahNumberEightyOne              LabelAyahNumber  = 2081
	LabelAyahNumberEightyTwo              LabelAyahNumber  = 2082
	LabelAyahNumberEightyThree            LabelAyahNumber  = 2083
	LabelAyahNumberEightyFour             LabelAyahNumber  = 2084
	LabelAyahNumberEightyFive             LabelAyahNumber  = 2085
	LabelAyahNumberEightySix              LabelAyahNumber  = 2086
	LabelAyahNumberEightySeven            LabelAyahNumber  = 2087
	LabelAyahNumberEightyEight            LabelAyahNumber  = 2088
	LabelAyahNumberEightyNine             LabelAyahNumber  = 2089
	LabelAyahNumberNinety                 LabelAyahNumber  = 2090
	LabelAyahNumberNinetyOne              LabelAyahNumber  = 2091
	LabelAyahNumberNinetyTwo              LabelAyahNumber  = 2092
	LabelAyahNumberNinetyThree            LabelAyahNumber  = 2093
	LabelAyahNumberNinetyFour             LabelAyahNumber  = 2094
	LabelAyahNumberNinetyFive             LabelAyahNumber  = 2095
	LabelAyahNumberNinetySix              LabelAyahNumber  = 2096
	LabelAyahNumberNinetySeven            LabelAyahNumber  = 2097
	LabelAyahNumberNinetyEight            LabelAyahNumber  = 2098
	LabelAyahNumberNinetyNine             LabelAyahNumber  = 2099
	LabelAyahNumberOneHundred             LabelAyahNumber  = 2100
	LabelAyahNumberOneHundredOne          LabelAyahNumber  = 2101
	LabelAyahNumberOneHundredTwo          LabelAyahNumber  = 2102
	LabelAyahNumberOneHundredThree        LabelAyahNumber  = 2103
	LabelAyahNumberOneHundredFour         LabelAyahNumber  = 2104
	LabelAyahNumberOneHundredFive         LabelAyahNumber  = 2105
	LabelAyahNumberOneHundredSix          LabelAyahNumber  = 2106
	LabelAyahNumberOneHundredSeven        LabelAyahNumber  = 2107
	LabelAyahNumberOneHundredEight        LabelAyahNumber  = 2108
	LabelAyahNumberOneHundredNine         LabelAyahNumber  = 2109
	LabelAyahNumberOneHundredTen          LabelAyahNumber  = 2110
	LabelAyahNumberOneHundredEleven       LabelAyahNumber  = 2111
	LabelAyahNumberOneHundredTwelve       LabelAyahNumber  = 2112
	LabelAyahNumberOneHundredThirteen     LabelAyahNumber  = 2113
	LabelAyahNumberOneHundredFourteen     LabelAyahNumber  = 2114
	LabelAyahNumberOneHundredFifteen      LabelAyahNumber  = 2115
	LabelAyahNumberOneHundredSixteen      LabelAyahNumber  = 2116
	LabelAyahNumberOneHundredSeventeen    LabelAyahNumber  = 2117
	LabelAyahNumberOneHundredEighteen     LabelAyahNumber  = 2118
	LabelAyahNumberOneHundredNineteen     LabelAyahNumber  = 2119
	LabelAyahNumberOneHundredTwenty       LabelAyahNumber  = 2120
	LabelAyahNumberOneHundredTwentyOne    LabelAyahNumber  = 2121
	LabelAyahNumberOneHundredTwentyTwo    LabelAyahNumber  = 2122
	LabelAyahNumberOneHundredTwentyThree  LabelAyahNumber  = 2123
	LabelAyahNumberOneHundredTwentyFour   LabelAyahNumber  = 2124
	LabelAyahNumberOneHundredTwentyFive   LabelAyahNumber  = 2125
	LabelAyahNumberOneHundredTwentySix    LabelAyahNumber  = 2126
	LabelAyahNumberOneHundredTwentySeven  LabelAyahNumber  = 2127
	LabelAyahNumberOneHundredTwentyEight  LabelAyahNumber  = 2128
	LabelAyahNumberOneHundredTwentyNine   LabelAyahNumber  = 2129
	LabelAyahNumberOneHundredThirty       LabelAyahNumber  = 2130
	LabelAyahNumberOneHundredThirtyOne    LabelAyahNumber  = 2131
	LabelAyahNumberOneHundredThirtyTwo    LabelAyahNumber  = 2132
	LabelAyahNumberOneHundredThirtyThree  LabelAyahNumber  = 2133
	LabelAyahNumberOneHundredThirtyFour   LabelAyahNumber  = 2134
	LabelAyahNumberOneHundredThirtyFive   LabelAyahNumber  = 2135
	LabelAyahNumberOneHundredThirtySix    LabelAyahNumber  = 2136
	LabelAyahNumberOneHundredThirtySeven  LabelAyahNumber  = 2137
	LabelAyahNumberOneHundredThirtyEight  LabelAyahNumber  = 2138
	LabelAyahNumberOneHundredThirtyNine   LabelAyahNumber  = 2139
	LabelAyahNumberOneHundredFourty       LabelAyahNumber  = 2140
	LabelAyahNumberOneHundredFourtyOne    LabelAyahNumber  = 2141
	LabelAyahNumberOneHundredFourtyTwo    LabelAyahNumber  = 2142
	LabelAyahNumberOneHundredFourtyThree  LabelAyahNumber  = 2143
	LabelAyahNumberOneHundredFourtyFour   LabelAyahNumber  = 2144
	LabelAyahNumberOneHundredFourtyFive   LabelAyahNumber  = 2145
	LabelAyahNumberOneHundredFourtySix    LabelAyahNumber  = 2146
	LabelAyahNumberOneHundredFourtySeven  LabelAyahNumber  = 2147
	LabelAyahNumberOneHundredFourtyEight  LabelAyahNumber  = 2148
	LabelAyahNumberOneHundredFourtyNine   LabelAyahNumber  = 2149
	LabelAyahNumberOneHundredFifty        LabelAyahNumber  = 2150
	LabelAyahNumberOneHundredFiftyOne     LabelAyahNumber  = 2151
	LabelAyahNumberOneHundredFiftyTwo     LabelAyahNumber  = 2152
	LabelAyahNumberOneHundredFiftyThree   LabelAyahNumber  = 2153
	LabelAyahNumberOneHundredFiftyFour    LabelAyahNumber  = 2154
	LabelAyahNumberOneHundredFiftyFive    LabelAyahNumber  = 2155
	LabelAyahNumberOneHundredFiftySix     LabelAyahNumber  = 2156
	LabelAyahNumberOneHundredFiftySeven   LabelAyahNumber  = 2157
	LabelAyahNumberOneHundredFiftyEight   LabelAyahNumber  = 2158
	LabelAyahNumberOneHundredFiftyNine    LabelAyahNumber  = 2159
	LabelAyahNumberOneHundredSixty        LabelAyahNumber  = 2160
	LabelAyahNumberOneHundredSixtyOne     LabelAyahNumber  = 2161
	LabelAyahNumberOneHundredSixtyTwo     LabelAyahNumber  = 2162
	LabelAyahNumberOneHundredSixtyThree   LabelAyahNumber  = 2163
	LabelAyahNumberOneHundredSixtyFour    LabelAyahNumber  = 2164
	LabelAyahNumberOneHundredSixtyFive    LabelAyahNumber  = 2165
	LabelAyahNumberOneHundredSixtySix     LabelAyahNumber  = 2166
	LabelAyahNumberOneHundredSixtySeven   LabelAyahNumber  = 2167
	LabelAyahNumberOneHundredSixtyEight   LabelAyahNumber  = 2168
	LabelAyahNumberOneHundredSixtyNine    LabelAyahNumber  = 2169
	LabelAyahNumberOneHundredSeventy      LabelAyahNumber  = 2170
	LabelAyahNumberOneHundredSeventyOne   LabelAyahNumber  = 2171
	LabelAyahNumberOneHundredSeventyTwo   LabelAyahNumber  = 2172
	LabelAyahNumberOneHundredSeventyThree LabelAyahNumber  = 2173
	LabelAyahNumberOneHundredSeventyFour  LabelAyahNumber  = 2174
	LabelAyahNumberOneHundredSeventyFive  LabelAyahNumber  = 2175
	LabelAyahNumberOneHundredSeventySix   LabelAyahNumber  = 2176
	LabelAyahNumberOneHundredSeventySeven LabelAyahNumber  = 2177
	LabelAyahNumberOneHundredSeventyEight LabelAyahNumber  = 2178
	LabelAyahNumberOneHundredSeventyNine  LabelAyahNumber  = 2179
	LabelAyahNumberOneHundredEighty       LabelAyahNumber  = 2180
	LabelAyahNumberOneHundredEightyOne    LabelAyahNumber  = 2181
	LabelAyahNumberOneHundredEightyTwo    LabelAyahNumber  = 2182
	LabelAyahNumberOneHundredEightyThree  LabelAyahNumber  = 2183
	LabelAyahNumberOneHundredEightyFour   LabelAyahNumber  = 2184
	LabelAyahNumberOneHundredEightyFive   LabelAyahNumber  = 2185
	LabelAyahNumberOneHundredEightySix    LabelAyahNumber  = 2186
	LabelAyahNumberOneHundredEightySeven  LabelAyahNumber  = 2187
	LabelAyahNumberOneHundredEightyEight  LabelAyahNumber  = 2188
	LabelAyahNumberOneHundredEightyNine   LabelAyahNumber  = 2189
	LabelAyahNumberOneHundredNinety       LabelAyahNumber  = 2190
	LabelAyahNumberOneHundredNinetyOne    LabelAyahNumber  = 2191
	LabelAyahNumberOneHundredNinetyTwo    LabelAyahNumber  = 2192
	LabelAyahNumberOneHundredNinetyThree  LabelAyahNumber  = 2193
	LabelAyahNumberOneHundredNinetyFour   LabelAyahNumber  = 2194
	LabelAyahNumberOneHundredNinetyFive   LabelAyahNumber  = 2195
	LabelAyahNumberOneHundredNinetySix    LabelAyahNumber  = 2196
	LabelAyahNumberOneHundredNinetySeven  LabelAyahNumber  = 2197
	LabelAyahNumberOneHundredNinetyEight  LabelAyahNumber  = 2198
	LabelAyahNumberOneHundredNinetyNine   LabelAyahNumber  = 2199
	LabelAyahNumberTwoHundred             LabelAyahNumber  = 2200
	LabelAyahNumberTwoHundredOne          LabelAyahNumber  = 2201
	LabelAyahNumberTwoHundredTwo          LabelAyahNumber  = 2202
	LabelAyahNumberTwoHundredThree        LabelAyahNumber  = 2203
	LabelAyahNumberTwoHundredFour         LabelAyahNumber  = 2204
	LabelAyahNumberTwoHundredFive         LabelAyahNumber  = 2205
	LabelAyahNumberTwoHundredSix          LabelAyahNumber  = 2206
	LabelAyahNumberTwoHundredSeven        LabelAyahNumber  = 2207
	LabelAyahNumberTwoHundredEight        LabelAyahNumber  = 2208
	LabelAyahNumberTwoHundredNine         LabelAyahNumber  = 2209
	LabelAyahNumberTwoHundredTen          LabelAyahNumber  = 2210
	LabelAyahNumberTwoHundredEleven       LabelAyahNumber  = 2211
	LabelAyahNumberTwoHundredTwelve       LabelAyahNumber  = 2212
	LabelAyahNumberTwoHundredThirteen     LabelAyahNumber  = 2213
	LabelAyahNumberTwoHundredFourteen     LabelAyahNumber  = 2214
	LabelAyahNumberTwoHundredFifteen      LabelAyahNumber  = 2215
	LabelAyahNumberTwoHundredSixteen      LabelAyahNumber  = 2216
	LabelAyahNumberTwoHundredSeventeen    LabelAyahNumber  = 2217
	LabelAyahNumberTwoHundredEighteen     LabelAyahNumber  = 2218
	LabelAyahNumberTwoHundredNineteen     LabelAyahNumber  = 2219
	LabelAyahNumberTwoHundredTwenty       LabelAyahNumber  = 2220
	LabelAyahNumberTwoHundredTwentyOne    LabelAyahNumber  = 2221
	LabelAyahNumberTwoHundredTwentyTwo    LabelAyahNumber  = 2222
	LabelAyahNumberTwoHundredTwentyThree  LabelAyahNumber  = 2223
	LabelAyahNumberTwoHundredTwentyFour   LabelAyahNumber  = 2224
	LabelAyahNumberTwoHundredTwentyFive   LabelAyahNumber  = 2225
	LabelAyahNumberTwoHundredTwentySix    LabelAyahNumber  = 2226
	LabelAyahNumberTwoHundredTwentySeven  LabelAyahNumber  = 2227
	LabelAyahNumberTwoHundredTwentyEight  LabelAyahNumber  = 2228
	LabelAyahNumberTwoHundredTwentyNine   LabelAyahNumber  = 2229
	LabelAyahNumberTwoHundredThirty       LabelAyahNumber  = 2230
	LabelAyahNumberTwoHundredThirtyOne    LabelAyahNumber  = 2231
	LabelAyahNumberTwoHundredThirtyTwo    LabelAyahNumber  = 2232
	LabelAyahNumberTwoHundredThirtyThree  LabelAyahNumber  = 2233
	LabelAyahNumberTwoHundredThirtyFour   LabelAyahNumber  = 2234
	LabelAyahNumberTwoHundredThirtyFive   LabelAyahNumber  = 2235
	LabelAyahNumberTwoHundredThirtySix    LabelAyahNumber  = 2236
	LabelAyahNumberTwoHundredThirtySeven  LabelAyahNumber  = 2237
	LabelAyahNumberTwoHundredThirtyEight  LabelAyahNumber  = 2238
	LabelAyahNumberTwoHundredThirtyNine   LabelAyahNumber  = 2239
	LabelAyahNumberTwoHundredFourty       LabelAyahNumber  = 2240
	LabelAyahNumberTwoHundredFourtyOne    LabelAyahNumber  = 2241
	LabelAyahNumberTwoHundredFourtyTwo    LabelAyahNumber  = 2242
	LabelAyahNumberTwoHundredFourtyThree  LabelAyahNumber  = 2243
	LabelAyahNumberTwoHundredFourtyFour   LabelAyahNumber  = 2244
	LabelAyahNumberTwoHundredFourtyFive   LabelAyahNumber  = 2245
	LabelAyahNumberTwoHundredFourtySix    LabelAyahNumber  = 2246
	LabelAyahNumberTwoHundredFourtySeven  LabelAyahNumber  = 2247
	LabelAyahNumberTwoHundredFourtyEight  LabelAyahNumber  = 2248
	LabelAyahNumberTwoHundredFourtyNine   LabelAyahNumber  = 2249
	LabelAyahNumberTwoHundredFifty        LabelAyahNumber  = 2250
	LabelAyahNumberTwoHundredFiftyOne     LabelAyahNumber  = 2251
	LabelAyahNumberTwoHundredFiftyTwo     LabelAyahNumber  = 2252
	LabelAyahNumberTwoHundredFiftyThree   LabelAyahNumber  = 2253
	LabelAyahNumberTwoHundredFiftyFour    LabelAyahNumber  = 2254
	LabelAyahNumberTwoHundredFiftyFive    LabelAyahNumber  = 2255
	LabelAyahNumberTwoHundredFiftySix     LabelAyahNumber  = 2256
	LabelAyahNumberTwoHundredFiftySeven   LabelAyahNumber  = 2257
	LabelAyahNumberTwoHundredFiftyEight   LabelAyahNumber  = 2258
	LabelAyahNumberTwoHundredFiftyNine    LabelAyahNumber  = 2259
	LabelAyahNumberTwoHundredSixty        LabelAyahNumber  = 2260
	LabelAyahNumberTwoHundredSixtyOne     LabelAyahNumber  = 2261
	LabelAyahNumberTwoHundredSixtyTwo     LabelAyahNumber  = 2262
	LabelAyahNumberTwoHundredSixtyThree   LabelAyahNumber  = 2263
	LabelAyahNumberTwoHundredSixtyFour    LabelAyahNumber  = 2264
	LabelAyahNumberTwoHundredSixtyFive    LabelAyahNumber  = 2265
	LabelAyahNumberTwoHundredSixtySix     LabelAyahNumber  = 2266
	LabelAyahNumberTwoHundredSixtySeven   LabelAyahNumber  = 2267
	LabelAyahNumberTwoHundredSixtyEight   LabelAyahNumber  = 2268
	LabelAyahNumberTwoHundredSixtyNine    LabelAyahNumber  = 2269
	LabelAyahNumberTwoHundredSeventy      LabelAyahNumber  = 2270
	LabelAyahNumberTwoHundredSeventyOne   LabelAyahNumber  = 2271
	LabelAyahNumberTwoHundredSeventyTwo   LabelAyahNumber  = 2272
	LabelAyahNumberTwoHundredSeventyThree LabelAyahNumber  = 2273
	LabelAyahNumberTwoHundredSeventyFour  LabelAyahNumber  = 2274
	LabelAyahNumberTwoHundredSeventyFive  LabelAyahNumber  = 2275
	LabelAyahNumberTwoHundredSeventySix   LabelAyahNumber  = 2276
	LabelAyahNumberTwoHundredSeventySeven LabelAyahNumber  = 2277
	LabelAyahNumberTwoHundredSeventyEight LabelAyahNumber  = 2278
	LabelAyahNumberTwoHundredSeventyNine  LabelAyahNumber  = 2279
	LabelAyahNumberTwoHundredEighty       LabelAyahNumber  = 2280
	LabelAyahNumberTwoHundredEightyOne    LabelAyahNumber  = 2281
	LabelAyahNumberTwoHundredEightyTwo    LabelAyahNumber  = 2282
	LabelAyahNumberTwoHundredEightyThree  LabelAyahNumber  = 2283
	LabelAyahNumberTwoHundredEightyFour   LabelAyahNumber  = 2284
	LabelAyahNumberTwoHundredEightyFive   LabelAyahNumber  = 2285
	LabelAyahNumberTwoHundredEightySix    LabelAyahNumber  = 2286
)

type LabelContentType int
type LabelSource int
type LabelSurahNumber int
type LabelAyahNumber int

var ContentTypeToLabel = map[database.ContentType]LabelContentType{
	database.ContentTypeTafsir: LabelContentTypeTafsir,
}

var SourceToLabel = map[database.Source]LabelSource{
	database.SourceTafsirIbnKathir:  LabelSourceTafsirIbnKathir,
	database.SourceTafsirAlTabari:   LabelSourceTafsirAlTabari,
	database.SourceTafsirAlQurtubi:  LabelSourceTafsirAlQurtubi,
	database.SourceTafsirAlBaghawi:  LabelSourceTafsirAlBaghawi,
	database.SourceTafsirAlSaadi:    LabelSourceTafsirAlSaadi,
	database.SourceTafsirAlMuyassar: LabelSourceTafsirAlMuyassar,
	database.SourceTafsirAlWasit:    LabelSourceTafsirAlWasit,
	database.SourceTafsirAlJalalayn: LabelSourceTafsirAlJalalayn,
}

var SurahNumberToLabel = map[int32]LabelSurahNumber{
	1:   LabelSurahNumberOne,
	2:   LabelSurahNumberTwo,
	3:   LabelSurahNumberThree,
	4:   LabelSurahNumberFour,
	5:   LabelSurahNumberFive,
	6:   LabelSurahNumberSix,
	7:   LabelSurahNumberSeven,
	8:   LabelSurahNumberEight,
	9:   LabelSurahNumberNine,
	10:  LabelSurahNumberTen,
	11:  LabelSurahNumberEleven,
	12:  LabelSurahNumberTwelve,
	13:  LabelSurahNumberThirteen,
	14:  LabelSurahNumberFourteen,
	15:  LabelSurahNumberFifteen,
	16:  LabelSurahNumberSixteen,
	17:  LabelSurahNumberSeventeen,
	18:  LabelSurahNumberEighteen,
	19:  LabelSurahNumberNineteen,
	20:  LabelSurahNumberTwenty,
	21:  LabelSurahNumberTwentyOne,
	22:  LabelSurahNumberTwentyTwo,
	23:  LabelSurahNumberTwentyThree,
	24:  LabelSurahNumberTwentyFour,
	25:  LabelSurahNumberTwentyFive,
	26:  LabelSurahNumberTwentySix,
	27:  LabelSurahNumberTwentySeven,
	28:  LabelSurahNumberTwentyEight,
	29:  LabelSurahNumberTwentyNine,
	30:  LabelSurahNumberThirty,
	31:  LabelSurahNumberThirtyOne,
	32:  LabelSurahNumberThirtyTwo,
	33:  LabelSurahNumberThirtyThree,
	34:  LabelSurahNumberThirtyFour,
	35:  LabelSurahNumberThirtyFive,
	36:  LabelSurahNumberThirtySix,
	37:  LabelSurahNumberThirtySeven,
	38:  LabelSurahNumberThirtyEight,
	39:  LabelSurahNumberThirtyNine,
	40:  LabelSurahNumberFourty,
	41:  LabelSurahNumberFourtyOne,
	42:  LabelSurahNumberFourtyTwo,
	43:  LabelSurahNumberFourtyThree,
	44:  LabelSurahNumberFourtyFour,
	45:  LabelSurahNumberFourtyFive,
	46:  LabelSurahNumberFourtySix,
	47:  LabelSurahNumberFourtySeven,
	48:  LabelSurahNumberFourtyEight,
	49:  LabelSurahNumberFourtyNine,
	50:  LabelSurahNumberFifty,
	51:  LabelSurahNumberFiftyOne,
	52:  LabelSurahNumberFiftyTwo,
	53:  LabelSurahNumberFiftyThree,
	54:  LabelSurahNumberFiftyFour,
	55:  LabelSurahNumberFiftyFive,
	56:  LabelSurahNumberFiftySix,
	57:  LabelSurahNumberFiftySeven,
	58:  LabelSurahNumberFiftyEight,
	59:  LabelSurahNumberFiftyNine,
	60:  LabelSurahNumberSixty,
	61:  LabelSurahNumberSixtyOne,
	62:  LabelSurahNumberSixtyTwo,
	63:  LabelSurahNumberSixtyThree,
	64:  LabelSurahNumberSixtyFour,
	65:  LabelSurahNumberSixtyFive,
	66:  LabelSurahNumberSixtySix,
	67:  LabelSurahNumberSixtySeven,
	68:  LabelSurahNumberSixtyEight,
	69:  LabelSurahNumberSixtyNine,
	70:  LabelSurahNumberSeventy,
	71:  LabelSurahNumberSeventyOne,
	72:  LabelSurahNumberSeventyTwo,
	73:  LabelSurahNumberSeventyThree,
	74:  LabelSurahNumberSeventyFour,
	75:  LabelSurahNumberSeventyFive,
	76:  LabelSurahNumberSeventySix,
	77:  LabelSurahNumberSeventySeven,
	78:  LabelSurahNumberSeventyEight,
	79:  LabelSurahNumberSeventyNine,
	80:  LabelSurahNumberEighty,
	81:  LabelSurahNumberEightyOne,
	82:  LabelSurahNumberEightyTwo,
	83:  LabelSurahNumberEightyThree,
	84:  LabelSurahNumberEightyFour,
	85:  LabelSurahNumberEightyFive,
	86:  LabelSurahNumberEightySix,
	87:  LabelSurahNumberEightySeven,
	88:  LabelSurahNumberEightyEight,
	89:  LabelSurahNumberEightyNine,
	90:  LabelSurahNumberNinety,
	91:  LabelSurahNumberNinetyOne,
	92:  LabelSurahNumberNinetyTwo,
	93:  LabelSurahNumberNinetyThree,
	94:  LabelSurahNumberNinetyFour,
	95:  LabelSurahNumberNinetyFive,
	96:  LabelSurahNumberNinetySix,
	97:  LabelSurahNumberNinetySeven,
	98:  LabelSurahNumberNinetyEight,
	99:  LabelSurahNumberNinetyNine,
	100: LabelSurahNumberOneHundred,
	101: LabelSurahNumberOneHundredOne,
	102: LabelSurahNumberOneHundredTwo,
	103: LabelSurahNumberOneHundredThree,
	104: LabelSurahNumberOneHundredFour,
	105: LabelSurahNumberOneHundredFive,
	106: LabelSurahNumberOneHundredSix,
	107: LabelSurahNumberOneHundredSeven,
	108: LabelSurahNumberOneHundredEight,
	109: LabelSurahNumberOneHundredNine,
	110: LabelSurahNumberOneHundredTen,
	111: LabelSurahNumberOneHundredEleven,
	112: LabelSurahNumberOneHundredTwelve,
	113: LabelSurahNumberOneHundredThirteen,
	114: LabelSurahNumberOneHundredFourteen,
}

var AyahNumberToLabel = map[int32]LabelAyahNumber{
	1:   LabelAyahNumberOne,
	2:   LabelAyahNumberTwo,
	3:   LabelAyahNumberThree,
	4:   LabelAyahNumberFour,
	5:   LabelAyahNumberFive,
	6:   LabelAyahNumberSix,
	7:   LabelAyahNumberSeven,
	8:   LabelAyahNumberEight,
	9:   LabelAyahNumberNine,
	10:  LabelAyahNumberTen,
	11:  LabelAyahNumberEleven,
	12:  LabelAyahNumberTwelve,
	13:  LabelAyahNumberThirteen,
	14:  LabelAyahNumberFourteen,
	15:  LabelAyahNumberFifteen,
	16:  LabelAyahNumberSixteen,
	17:  LabelAyahNumberSeventeen,
	18:  LabelAyahNumberEighteen,
	19:  LabelAyahNumberNineteen,
	20:  LabelAyahNumberTwenty,
	21:  LabelAyahNumberTwentyOne,
	22:  LabelAyahNumberTwentyTwo,
	23:  LabelAyahNumberTwentyThree,
	24:  LabelAyahNumberTwentyFour,
	25:  LabelAyahNumberTwentyFive,
	26:  LabelAyahNumberTwentySix,
	27:  LabelAyahNumberTwentySeven,
	28:  LabelAyahNumberTwentyEight,
	29:  LabelAyahNumberTwentyNine,
	30:  LabelAyahNumberThirty,
	31:  LabelAyahNumberThirtyOne,
	32:  LabelAyahNumberThirtyTwo,
	33:  LabelAyahNumberThirtyThree,
	34:  LabelAyahNumberThirtyFour,
	35:  LabelAyahNumberThirtyFive,
	36:  LabelAyahNumberThirtySix,
	37:  LabelAyahNumberThirtySeven,
	38:  LabelAyahNumberThirtyEight,
	39:  LabelAyahNumberThirtyNine,
	40:  LabelAyahNumberFourty,
	41:  LabelAyahNumberFourtyOne,
	42:  LabelAyahNumberFourtyTwo,
	43:  LabelAyahNumberFourtyThree,
	44:  LabelAyahNumberFourtyFour,
	45:  LabelAyahNumberFourtyFive,
	46:  LabelAyahNumberFourtySix,
	47:  LabelAyahNumberFourtySeven,
	48:  LabelAyahNumberFourtyEight,
	49:  LabelAyahNumberFourtyNine,
	50:  LabelAyahNumberFifty,
	51:  LabelAyahNumberFiftyOne,
	52:  LabelAyahNumberFiftyTwo,
	53:  LabelAyahNumberFiftyThree,
	54:  LabelAyahNumberFiftyFour,
	55:  LabelAyahNumberFiftyFive,
	56:  LabelAyahNumberFiftySix,
	57:  LabelAyahNumberFiftySeven,
	58:  LabelAyahNumberFiftyEight,
	59:  LabelAyahNumberFiftyNine,
	60:  LabelAyahNumberSixty,
	61:  LabelAyahNumberSixtyOne,
	62:  LabelAyahNumberSixtyTwo,
	63:  LabelAyahNumberSixtyThree,
	64:  LabelAyahNumberSixtyFour,
	65:  LabelAyahNumberSixtyFive,
	66:  LabelAyahNumberSixtySix,
	67:  LabelAyahNumberSixtySeven,
	68:  LabelAyahNumberSixtyEight,
	69:  LabelAyahNumberSixtyNine,
	70:  LabelAyahNumberSeventy,
	71:  LabelAyahNumberSeventyOne,
	72:  LabelAyahNumberSeventyTwo,
	73:  LabelAyahNumberSeventyThree,
	74:  LabelAyahNumberSeventyFour,
	75:  LabelAyahNumberSeventyFive,
	76:  LabelAyahNumberSeventySix,
	77:  LabelAyahNumberSeventySeven,
	78:  LabelAyahNumberSeventyEight,
	79:  LabelAyahNumberSeventyNine,
	80:  LabelAyahNumberEighty,
	81:  LabelAyahNumberEightyOne,
	82:  LabelAyahNumberEightyTwo,
	83:  LabelAyahNumberEightyThree,
	84:  LabelAyahNumberEightyFour,
	85:  LabelAyahNumberEightyFive,
	86:  LabelAyahNumberEightySix,
	87:  LabelAyahNumberEightySeven,
	88:  LabelAyahNumberEightyEight,
	89:  LabelAyahNumberEightyNine,
	90:  LabelAyahNumberNinety,
	91:  LabelAyahNumberNinetyOne,
	92:  LabelAyahNumberNinetyTwo,
	93:  LabelAyahNumberNinetyThree,
	94:  LabelAyahNumberNinetyFour,
	95:  LabelAyahNumberNinetyFive,
	96:  LabelAyahNumberNinetySix,
	97:  LabelAyahNumberNinetySeven,
	98:  LabelAyahNumberNinetyEight,
	99:  LabelAyahNumberNinetyNine,
	100: LabelAyahNumberOneHundred,
	101: LabelAyahNumberOneHundredOne,
	102: LabelAyahNumberOneHundredTwo,
	103: LabelAyahNumberOneHundredThree,
	104: LabelAyahNumberOneHundredFour,
	105: LabelAyahNumberOneHundredFive,
	106: LabelAyahNumberOneHundredSix,
	107: LabelAyahNumberOneHundredSeven,
	108: LabelAyahNumberOneHundredEight,
	109: LabelAyahNumberOneHundredNine,
	110: LabelAyahNumberOneHundredTen,
	111: LabelAyahNumberOneHundredEleven,
	112: LabelAyahNumberOneHundredTwelve,
	113: LabelAyahNumberOneHundredThirteen,
	114: LabelAyahNumberOneHundredFourteen,
	115: LabelAyahNumberOneHundredFifteen,
	116: LabelAyahNumberOneHundredSixteen,
	117: LabelAyahNumberOneHundredSeventeen,
	118: LabelAyahNumberOneHundredEighteen,
	119: LabelAyahNumberOneHundredNineteen,
	120: LabelAyahNumberOneHundredTwenty,
	121: LabelAyahNumberOneHundredTwentyOne,
	122: LabelAyahNumberOneHundredTwentyTwo,
	123: LabelAyahNumberOneHundredTwentyThree,
	124: LabelAyahNumberOneHundredTwentyFour,
	125: LabelAyahNumberOneHundredTwentyFive,
	126: LabelAyahNumberOneHundredTwentySix,
	127: LabelAyahNumberOneHundredTwentySeven,
	128: LabelAyahNumberOneHundredTwentyEight,
	129: LabelAyahNumberOneHundredTwentyNine,
	130: LabelAyahNumberOneHundredThirty,
	131: LabelAyahNumberOneHundredThirtyOne,
	132: LabelAyahNumberOneHundredThirtyTwo,
	133: LabelAyahNumberOneHundredThirtyThree,
	134: LabelAyahNumberOneHundredThirtyFour,
	135: LabelAyahNumberOneHundredThirtyFive,
	136: LabelAyahNumberOneHundredThirtySix,
	137: LabelAyahNumberOneHundredThirtySeven,
	138: LabelAyahNumberOneHundredThirtyEight,
	139: LabelAyahNumberOneHundredThirtyNine,
	140: LabelAyahNumberOneHundredFourty,
	141: LabelAyahNumberOneHundredFourtyOne,
	142: LabelAyahNumberOneHundredFourtyTwo,
	143: LabelAyahNumberOneHundredFourtyThree,
	144: LabelAyahNumberOneHundredFourtyFour,
	145: LabelAyahNumberOneHundredFourtyFive,
	146: LabelAyahNumberOneHundredFourtySix,
	147: LabelAyahNumberOneHundredFourtySeven,
	148: LabelAyahNumberOneHundredFourtyEight,
	149: LabelAyahNumberOneHundredFourtyNine,
	150: LabelAyahNumberOneHundredFifty,
	151: LabelAyahNumberOneHundredFiftyOne,
	152: LabelAyahNumberOneHundredFiftyTwo,
	153: LabelAyahNumberOneHundredFiftyThree,
	154: LabelAyahNumberOneHundredFiftyFour,
	155: LabelAyahNumberOneHundredFiftyFive,
	156: LabelAyahNumberOneHundredFiftySix,
	157: LabelAyahNumberOneHundredFiftySeven,
	158: LabelAyahNumberOneHundredFiftyEight,
	159: LabelAyahNumberOneHundredFiftyNine,
	160: LabelAyahNumberOneHundredSixty,
	161: LabelAyahNumberOneHundredSixtyOne,
	162: LabelAyahNumberOneHundredSixtyTwo,
	163: LabelAyahNumberOneHundredSixtyThree,
	164: LabelAyahNumberOneHundredSixtyFour,
	165: LabelAyahNumberOneHundredSixtyFive,
	166: LabelAyahNumberOneHundredSixtySix,
	167: LabelAyahNumberOneHundredSixtySeven,
	168: LabelAyahNumberOneHundredSixtyEight,
	169: LabelAyahNumberOneHundredSixtyNine,
	170: LabelAyahNumberOneHundredSeventy,
	171: LabelAyahNumberOneHundredSeventyOne,
	172: LabelAyahNumberOneHundredSeventyTwo,
	173: LabelAyahNumberOneHundredSeventyThree,
	174: LabelAyahNumberOneHundredSeventyFour,
	175: LabelAyahNumberOneHundredSeventyFive,
	176: LabelAyahNumberOneHundredSeventySix,
	177: LabelAyahNumberOneHundredSeventySeven,
	178: LabelAyahNumberOneHundredSeventyEight,
	179: LabelAyahNumberOneHundredSeventyNine,
	180: LabelAyahNumberOneHundredEighty,
	181: LabelAyahNumberOneHundredEightyOne,
	182: LabelAyahNumberOneHundredEightyTwo,
	183: LabelAyahNumberOneHundredEightyThree,
	184: LabelAyahNumberOneHundredEightyFour,
	185: LabelAyahNumberOneHundredEightyFive,
	186: LabelAyahNumberOneHundredEightySix,
	187: LabelAyahNumberOneHundredEightySeven,
	188: LabelAyahNumberOneHundredEightyEight,
	189: LabelAyahNumberOneHundredEightyNine,
	190: LabelAyahNumberOneHundredNinety,
	191: LabelAyahNumberOneHundredNinetyOne,
	192: LabelAyahNumberOneHundredNinetyTwo,
	193: LabelAyahNumberOneHundredNinetyThree,
	194: LabelAyahNumberOneHundredNinetyFour,
	195: LabelAyahNumberOneHundredNinetyFive,
	196: LabelAyahNumberOneHundredNinetySix,
	197: LabelAyahNumberOneHundredNinetySeven,
	198: LabelAyahNumberOneHundredNinetyEight,
	199: LabelAyahNumberOneHundredNinetyNine,
	200: LabelAyahNumberTwoHundred,
	201: LabelAyahNumberTwoHundredOne,
	202: LabelAyahNumberTwoHundredTwo,
	203: LabelAyahNumberTwoHundredThree,
	204: LabelAyahNumberTwoHundredFour,
	205: LabelAyahNumberTwoHundredFive,
	206: LabelAyahNumberTwoHundredSix,
	207: LabelAyahNumberTwoHundredSeven,
	208: LabelAyahNumberTwoHundredEight,
	209: LabelAyahNumberTwoHundredNine,
	210: LabelAyahNumberTwoHundredTen,
	211: LabelAyahNumberTwoHundredEleven,
	212: LabelAyahNumberTwoHundredTwelve,
	213: LabelAyahNumberTwoHundredThirteen,
	214: LabelAyahNumberTwoHundredFourteen,
	215: LabelAyahNumberTwoHundredFifteen,
	216: LabelAyahNumberTwoHundredSixteen,
	217: LabelAyahNumberTwoHundredSeventeen,
	218: LabelAyahNumberTwoHundredEighteen,
	219: LabelAyahNumberTwoHundredNineteen,
	220: LabelAyahNumberTwoHundredTwenty,
	221: LabelAyahNumberTwoHundredTwentyOne,
	222: LabelAyahNumberTwoHundredTwentyTwo,
	223: LabelAyahNumberTwoHundredTwentyThree,
	224: LabelAyahNumberTwoHundredTwentyFour,
	225: LabelAyahNumberTwoHundredTwentyFive,
	226: LabelAyahNumberTwoHundredTwentySix,
	227: LabelAyahNumberTwoHundredTwentySeven,
	228: LabelAyahNumberTwoHundredTwentyEight,
	229: LabelAyahNumberTwoHundredTwentyNine,
	230: LabelAyahNumberTwoHundredThirty,
	231: LabelAyahNumberTwoHundredThirtyOne,
	232: LabelAyahNumberTwoHundredThirtyTwo,
	233: LabelAyahNumberTwoHundredThirtyThree,
	234: LabelAyahNumberTwoHundredThirtyFour,
	235: LabelAyahNumberTwoHundredThirtyFive,
	236: LabelAyahNumberTwoHundredThirtySix,
	237: LabelAyahNumberTwoHundredThirtySeven,
	238: LabelAyahNumberTwoHundredThirtyEight,
	239: LabelAyahNumberTwoHundredThirtyNine,
	240: LabelAyahNumberTwoHundredFourty,
	241: LabelAyahNumberTwoHundredFourtyOne,
	242: LabelAyahNumberTwoHundredFourtyTwo,
	243: LabelAyahNumberTwoHundredFourtyThree,
	244: LabelAyahNumberTwoHundredFourtyFour,
	245: LabelAyahNumberTwoHundredFourtyFive,
	246: LabelAyahNumberTwoHundredFourtySix,
	247: LabelAyahNumberTwoHundredFourtySeven,
	248: LabelAyahNumberTwoHundredFourtyEight,
	249: LabelAyahNumberTwoHundredFourtyNine,
	250: LabelAyahNumberTwoHundredFifty,
	251: LabelAyahNumberTwoHundredFiftyOne,
	252: LabelAyahNumberTwoHundredFiftyTwo,
	253: LabelAyahNumberTwoHundredFiftyThree,
	254: LabelAyahNumberTwoHundredFiftyFour,
	255: LabelAyahNumberTwoHundredFiftyFive,
	256: LabelAyahNumberTwoHundredFiftySix,
	257: LabelAyahNumberTwoHundredFiftySeven,
	258: LabelAyahNumberTwoHundredFiftyEight,
	259: LabelAyahNumberTwoHundredFiftyNine,
	260: LabelAyahNumberTwoHundredSixty,
	261: LabelAyahNumberTwoHundredSixtyOne,
	262: LabelAyahNumberTwoHundredSixtyTwo,
	263: LabelAyahNumberTwoHundredSixtyThree,
	264: LabelAyahNumberTwoHundredSixtyFour,
	265: LabelAyahNumberTwoHundredSixtyFive,
	266: LabelAyahNumberTwoHundredSixtySix,
	267: LabelAyahNumberTwoHundredSixtySeven,
	268: LabelAyahNumberTwoHundredSixtyEight,
	269: LabelAyahNumberTwoHundredSixtyNine,
	270: LabelAyahNumberTwoHundredSeventy,
	271: LabelAyahNumberTwoHundredSeventyOne,
	272: LabelAyahNumberTwoHundredSeventyTwo,
	273: LabelAyahNumberTwoHundredSeventyThree,
	274: LabelAyahNumberTwoHundredSeventyFour,
	275: LabelAyahNumberTwoHundredSeventyFive,
	276: LabelAyahNumberTwoHundredSeventySix,
	277: LabelAyahNumberTwoHundredSeventySeven,
	278: LabelAyahNumberTwoHundredSeventyEight,
	279: LabelAyahNumberTwoHundredSeventyNine,
	280: LabelAyahNumberTwoHundredEighty,
	281: LabelAyahNumberTwoHundredEightyOne,
	282: LabelAyahNumberTwoHundredEightyTwo,
	283: LabelAyahNumberTwoHundredEightyThree,
	284: LabelAyahNumberTwoHundredEightyFour,
	285: LabelAyahNumberTwoHundredEightyFive,
	286: LabelAyahNumberTwoHundredEightySix,
}

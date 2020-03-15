package main

import (
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
)

type SteamResponse struct {
	Steamid string
	Success int
}

type Steam struct {
	Response SteamResponse
}

type Profile struct {
	Personaname string
	Avatarfull  string
}

type DotaProfile struct {
	Profile Profile
}

type DotaWinrate struct {
	Win  int
	Lose int
}

type DotaHero struct {
	Hero_id     string
	Last_played int
}

type DotaMatch struct {
	Match_id     int
	Kills        int
	Deaths       int
	Assists      int
	Hero_id      int
	Gold_per_min int
	Last_hits    int
}

type GameList struct {
	Results []GameObject
}

type GameObject struct {
	Id                    int
	Name                  string
	Original_release_date string
	Image                 GameImage
	Platforms             GamePlatform
}

type GameImage struct {
	Small_url string
}

type GamePlatform struct {
	Name         string
	Abbreviation string
}

func getData(url string) []byte {

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
	return body
}

func Request(url string, contentType string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(body)
	return body
}

func Rawurlencode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

func convert32bit(id_32 string) string {
	a, _ := new(big.Int).SetString(id_32, 10)
	b, _ := new(big.Int).SetString("76561197960265728", 10)
	return big.NewInt(0).Sub(a, b).Text(10)
}

func hero_id_to_names(id string) string {

	hero := make(map[string]string)

	hero["1"] = "Anti-Mage"
	hero["2"] = "Axe"
	hero["3"] = "Bane"
	hero["4"] = "Bloodseeker"
	hero["5"] = "Crystal Maiden"
	hero["6"] = "Drow"
	hero["7"] = "Earthshaker"
	hero["8"] = "Juggernaut"
	hero["9"] = "Mirana"
	hero["10"] = "Morphling"
	hero["11"] = "Shadow Fiend"
	hero["12"] = "Phantom Lancer"
	hero["13"] = "Puck"
	hero["14"] = "Pudge"
	hero["15"] = "Razor"
	hero["16"] = "Sand King"
	hero["17"] = "Storm Spirit"
	hero["18"] = "Sven"
	hero["19"] = "Tiny"
	hero["20"] = "Vengeful Spirit"
	hero["21"] = "Windranger"
	hero["22"] = "Zeus"
	hero["23"] = "Kunkka"
	hero["24"] = "Benin"
	hero["25"] = "Lina"
	hero["26"] = "Lion"
	hero["27"] = "Shadow Shaman"
	hero["28"] = "Slardar"
	hero["29"] = "Tidehunter"
	hero["30"] = "Witch Doctor"
	hero["31"] = "Lich"
	hero["32"] = "Riki"
	hero["33"] = "Enigma"
	hero["34"] = "Tinker"
	hero["35"] = "Sniper"
	hero["36"] = "Necrophos"
	hero["37"] = "Warlock"
	hero["38"] = "Beastmaster"
	hero["39"] = "Queen of Pain"
	hero["40"] = "Venomancer"
	hero["41"] = "Faceless Void"
	hero["42"] = "Wraith King"
	hero["43"] = "Death Prophet"
	hero["44"] = "Phantom Assasin"
	hero["45"] = "Pugna"
	hero["46"] = "Templar Assasin"
	hero["47"] = "Viper"
	hero["48"] = "Luna"
	hero["49"] = "Dragon Knight"
	hero["50"] = "Dazzle"
	hero["51"] = "Clockwerk"
	hero["52"] = "Leshrac"
	hero["53"] = "Nature Prohet"
	hero["54"] = "Lifestealer"
	hero["55"] = "Dark Seer"
	hero["56"] = "Clinkz"
	hero["57"] = "Omniknight"
	hero["58"] = "Enchantress"
	hero["59"] = "Huskar"
	hero["60"] = "Night Stalker"
	hero["61"] = "Broodmother"
	hero["62"] = "Bounty Hunter"
	hero["63"] = "Weaver"
	hero["64"] = "Jakiro"
	hero["65"] = "Batrider"
	hero["66"] = "Chen"
	hero["67"] = "Spectre"
	hero["68"] = "Ancient Apparition"
	hero["69"] = "Doom"
	hero["70"] = "Ursa"
	hero["71"] = "Spirit Breaker"
	hero["72"] = "Gyrocopter"
	hero["73"] = "Alchemist"
	hero["74"] = "Invoker"
	hero["75"] = "Silencer"
	hero["76"] = "Outworld Devourer"
	hero["77"] = "Lycan"
	hero["78"] = "Brewmaster"
	hero["79"] = "Shadow Demon"
	hero["80"] = "Lone Druid"
	hero["81"] = "Chaos Knight"
	hero["82"] = "Meepo"
	hero["83"] = "Treant Protector"
	hero["84"] = "Ogre Magi"
	hero["85"] = "Undying"
	hero["86"] = "Rubick"
	hero["87"] = "Disruptor"
	hero["88"] = "Nyx Assasin"
	hero["89"] = "Naga Siren"
	hero["90"] = "Keeper of the Light"
	hero["91"] = "Io"
	hero["92"] = "Visage"
	hero["93"] = "Slark"
	hero["94"] = "Medusa"
	hero["95"] = "Troll Warlord"
	hero["96"] = "Centaur Warrunner"
	hero["97"] = "Magnus"
	hero["98"] = "Timbersaw"
	hero["99"] = "Bristleback"
	hero["100"] = "Tusk"
	hero["101"] = "Skywrath Mage"
	hero["102"] = "Abaddon"
	hero["103"] = "Elder Titan"
	hero["104"] = "Legion Commander"
	hero["105"] = "Techies"
	hero["106"] = "Ember Spirit"
	hero["107"] = "Earth Spirit"
	hero["108"] = "Underlord"
	hero["109"] = "Terrorblade"
	hero["110"] = "Phoenix"
	hero["111"] = "Oracle"
	hero["112"] = "Winter Wyvern"
	hero["113"] = "Arc Warden"
	hero["114"] = "Monkey King"
	hero["119"] = "Dark Willow"
	hero["120"] = "Pangolier"

	return hero[id]
}

// func FlexGameJson(array []string) string {
// 	getArray := array
// 	dataCount := len(getArray)
// 	for i := 0; i <= dataCount; i++ {
// 	}
// }

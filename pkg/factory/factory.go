package factory

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"
)

var (
	firstNames = []string{
		"Budi", "Siti", "Eko", "Dewi", "Ilham", "Rini", "Agus", "Slamet",
		"Sri", "Bambang", "Rina", "Tri", "Hendra", "Nur", "Wawan", "Dian",
		"Bayu", "Maya", "Rizky", "Anisa", "Adi", "Mega", "Fajar", "Fitri",
		"Dimas", "Ayu", "Yanto", "Lestari", "Gita", "Teguh", "Indah", "Aris",
	}

	lastNames = []string{
		"Santoso", "Aminah", "Prasetyo", "Lestari", "Pratama", "Astuti",
		"Setiawan", "Riyadi", "Wahyuni", "Pamungkas", "Kusuma", "Hidayat",
		"Saputra", "Wibowo", "Utami", "Wijaya", "Siregar", "Nasution",
		"Firmansyah", "Purnama", "Subagyo", "Gunawan", "Hartono", "Manggala",
	}

	streets = []string{
		"Jl. Jenderal Sudirman", "Jl. Malioboro", "Jl. Gatot Subroto",
		"Jl. Diponegoro", "Jl. Pahlawan", "Jl. Ahmad Yani", "Jl. Pemuda",
		"Jl. Dr. Angka", "Jl. HR. Bunyamin", "Jl. Gerilya", "Jl. Overste Isdiman",
	}

	cities = []string{
		"Purwokerto", "Banyumas", "Cilacap", "Purbalingga", "Kebumen",
		"Banjarnegara", "Semarang", "Yogyakarta", "Solo", "Bandung",
		"Surabaya", "Jakarta", "Malang", "Denpasar",
	}

	words = []string{
		"larapak", "ngapak", "aplikasi", "database", "rute", "server", "cepet",
		"keplek", "mandiri", "apik", "aman", "modern", "banyumasan", "komponen",
		"layanan", "kreatif", "koneksi", "produktivitas", "skrip", "interaktif",
		"desain", "studio", "proyek", "digital", "inovasi", "kualitas", "lancar",
	}
)

// Factory provides pseudo-random realistic fake data generator using Go Standard Library.
type Factory struct{}

// New creates a new Factory generator instance.
func New() *Factory {
	return &Factory{}
}

func randInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return int(time.Now().UnixNano() % int64(max))
	}
	return int(n.Int64())
}

// Name generates a full person name.
func (f *Factory) Name() string {
	first := firstNames[randInt(len(firstNames))]
	last := lastNames[randInt(len(lastNames))]
	return first + " " + last
}

// Jeneng is the Ngapak alias for Name.
func (f *Factory) Jeneng() string {
	return f.Name()
}

// Email generates a realistic email address.
func (f *Factory) Email(name ...string) string {
	var prefix string
	if len(name) > 0 && name[0] != "" {
		clean := strings.ToLower(strings.ReplaceAll(name[0], " ", "."))
		prefix = clean
	} else {
		first := strings.ToLower(firstNames[randInt(len(firstNames))])
		last := strings.ToLower(lastNames[randInt(len(lastNames))])
		prefix = fmt.Sprintf("%s.%s%d", first, last, randInt(99)+1)
	}
	domains := []string{"example.com", "larapak.test", "ngapak.org", "pos.id"}
	dom := domains[randInt(len(domains))]
	return fmt.Sprintf("%s@%s", prefix, dom)
}

// Phone generates an Indonesian cellular phone number (e.g. 081234567890).
func (f *Factory) Phone() string {
	prefixes := []string{"0812", "0813", "0821", "0857", "0858", "0877", "0878", "0896"}
	pre := prefixes[randInt(len(prefixes))]
	num := randInt(90000000) + 10000000
	return fmt.Sprintf("%s%d", pre, num)
}

// HP is the Ngapak alias for Phone.
func (f *Factory) HP() string {
	return f.Phone()
}

// Address generates a street address with street number and city.
func (f *Factory) Address() string {
	st := streets[randInt(len(streets))]
	num := randInt(180) + 1
	city := cities[randInt(len(cities))]
	return fmt.Sprintf("%s No. %d, %s", st, num, city)
}

// Alamat is the Ngapak alias for Address.
func (f *Factory) Alamat() string {
	return f.Address()
}

// City generates a city name.
func (f *Factory) City() string {
	return cities[randInt(len(cities))]
}

// Kutha is the Ngapak alias for City.
func (f *Factory) Kutha() string {
	return f.City()
}

// Text generates a sentence containing a specific number of words.
func (f *Factory) Text(wordCount int) string {
	if wordCount <= 0 {
		wordCount = 10
	}
	var res []string
	for i := 0; i < wordCount; i++ {
		w := words[randInt(len(words))]
		if i == 0 {
			w = strings.Title(w)
		}
		res = append(res, w)
	}
	return strings.Join(res, " ") + "."
}

// Teks is the Ngapak alias for Text.
func (f *Factory) Teks(wordCount int) string {
	return f.Text(wordCount)
}

// Number generates an integer between min and max (inclusive).
func (f *Factory) Number(min, max int) int {
	if min >= max {
		return min
	}
	diff := max - min + 1
	return min + randInt(diff)
}

// Angka is the Ngapak alias for Number.
func (f *Factory) Angka(min, max int) int {
	return f.Number(min, max)
}

// Price generates a currency amount rounded to thousands.
func (f *Factory) Price(min, max int) int {
	if min < 1000 {
		min = 1000
	}
	if max <= min {
		max = min + 100000
	}
	steps := (max - min) / 1000
	return min + (randInt(steps+1) * 1000)
}

// Rega is the Ngapak alias for Price.
func (f *Factory) Rega(min, max int) int {
	return f.Price(min, max)
}

// Boolean generates a random boolean.
func (f *Factory) Boolean() bool {
	return randInt(2) == 1
}

// BenerSalah is the Ngapak alias for Boolean.
func (f *Factory) BenerSalah() bool {
	return f.Boolean()
}

// Date returns a formatted date string (YYYY-MM-DD) within the last two years.
func (f *Factory) Date() string {
	now := time.Now()
	daysAgo := randInt(730)
	target := now.AddDate(0, 0, -daysAgo)
	return target.Format("2006-01-02")
}

// Tanggal is the Ngapak alias for Date.
func (f *Factory) Tanggal() string {
	return f.Date()
}

// Choice selects a random item from a given slice.
func (f *Factory) Choice(items []any) any {
	if len(items) == 0 {
		return nil
	}
	return items[randInt(len(items))]
}

// Pilihan is the Ngapak alias for Choice.
func (f *Factory) Pilihan(items []any) any {
	return f.Choice(items)
}

// UUID generates a pseudo-UUID v4 string.
func (f *Factory) UUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Default returns a global default Factory instance.
var defaultFactory = New()

// Default returns the package default Factory.
func Default() *Factory {
	return defaultFactory
}

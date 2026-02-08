package auth

// Login credentials for FusionSolar
const (
	Username = "om@raitek.vn"
	Password = "Raitek@2024"
	LoginURL = "https://intl.fusionsolar.huawei.com/pvmswebsite/login/build/index.html#https%3A%2F%2Fintl.fusionsolar.huawei.com%2Funiportal%2Fpvmswebsite%2Fassets%2Fbuild%2Fcloud.html%3Fapp-id%3Dsmartpvms%26instance-id%3Dsmartpvms%26zone-id%3Dregion-7-5753e974-1570-4504-a020-d91a6d371d4c%23%2Fview%2Fdevice%2FNE%3D50987774%2Fmeter%2Fdetails"
)

// Credentials returns login info
func Credentials() (username, password, loginURL string) {
	return Username, Password, LoginURL
}

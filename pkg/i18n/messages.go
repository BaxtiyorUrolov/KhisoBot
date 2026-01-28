// pkg/i18n/messages.go
package i18n

type Messages struct {
	Welcome           string
	AskFullName       string
	InvalidFullName   string
	AskLocation       string
	InvalidLocation   string
	AskGrade          string
	InvalidGrade      string
	AskPhone          string
	AskOTP            string
	InvalidPhone      string
	InvalidOTP        string
	OTPSent           string
	RegistrationDone  string
	MainMenu          string
	AlreadyRegistered string
	Error             string
	ResendOTP         string
	NotRegistered     string
	BtnLogin          string
	BtnRegister       string
	BtnShareContact   string
	MustSubscribe     string
	BtnCheckSub       string
	SubscribeSuccess  string
}

var messages = map[string]Messages{
	"uz": {
		Welcome:           "👋 Xush kelibsiz!\n\nRo'yxatdan o'tish uchun ma'lumotlaringizni kiriting.",
		AskFullName:       "👤 Ism va familiyangizni kiriting:\n\n<i>Misol: Anvar Karimov</i>",
		InvalidFullName:   "❌ Ism va familiyangizni to'liq kiriting.\n\n<i>Misol: Anvar Karimov</i>",
		AskLocation:       "📍 Viloyat, tuman va maktabingizni kiriting:\n\n<i>Misol: Toshkent, Yunusobod, 56-maktab</i>",
		InvalidLocation:   "❌ Noto'g'ri format.\n\nQuyidagi formatda kiriting:\n<i>Viloyat, Tuman, Maktab</i>\n\nMisol: Toshkent, Yunusobod, 56-maktab",
		AskGrade:          "🎓 Nechanchi sinfda o'qiysiz?\n\n<i>1 dan 11 gacha raqam kiriting</i>",
		InvalidGrade:      "❌ Noto'g'ri sinf raqami.\n\n<i>1 dan 11 gacha raqam kiriting</i>",
		AskPhone:          "📱 Telefon raqamingizni kiriting yoki pastdagi tugmani bosing:\n\n<i>Misol: 998901234567</i>",
		AskOTP:            "🔐 Telefon raqamingizga yuborilgan tasdiqlash kodini kiriting:",
		InvalidPhone:      "❌ Telefon raqam noto'g'ri formatda.\n\n<i>Misol: 998901234567</i>",
		InvalidOTP:        "❌ Tasdiqlash kodi noto'g'ri. Qaytadan urinib ko'ring.",
		OTPSent:           "✅ Tasdiqlash kodi yuborildi: <b>%s</b>",
		RegistrationDone:  "🎉 Tabriklaymiz! Ro'yxatdan muvaffaqiyatli o'tdingiz.\n\n👤 Ism: <b>%s</b>\n👤 Familiya: <b>%s</b>\n🏙 Viloyat: <b>%s</b>\n🏘 Tuman: <b>%s</b>\n🏫 Maktab: <b>%s</b>\n🎓 Sinf: <b>%d</b>\n📱 Telefon: <b>%s</b>",
		MainMenu:          "Assalomu alaykum! Botimizga xush kelibsiz.\n\nAgar sizda <a href=\"https://khiso.uz\">khiso.uz</a> onlayn olimpiadalar platformasida akkaunt mavjud bo'lsa, \"Akkauntga kirish\" tugmasini bosing.\nAgar siz <a href=\"https://khiso.uz\">khiso.uz</a> onlayn olimpiadalar platformasidan ro'yxatdan o'tmagan bo'lsangiz, \"Akkaunt yaratish\" tugmasi orqali ro'yxatdan o'ting.\n\n<b>Diqqat!</b> Account yaratilgandan so'ng Olimpiadalar bo'limiga o'tib ro'yxatdan o'tishingiz mumkin.",
		AlreadyRegistered: "✅ Siz allaqachon ro'yxatdan o'tgansiz.",
		Error:             "❌ Xatolik yuz berdi. Iltimos qaytadan urinib ko'ring.",
		ResendOTP:         "🔄 Kodni qayta yuborish",
		NotRegistered:     "❌ Siz hali ro'yxatdan o'tmagansiz. /start bosing.",
		BtnLogin:          "🔑 Akkauntga kirish",
		BtnRegister:       "📝 Akkaunt yaratish",
		BtnShareContact:   "📱 Telefon raqamni ulashish",
		MustSubscribe:     "📢 Botdan foydalanish uchun quyidagi kanallarga obuna bo'ling:",
		BtnCheckSub:       "✅ Obunani tekshirish",
		SubscribeSuccess:  "✅ Rahmat! Endi botdan foydalanishingiz mumkin.",
	},
	"ru": {
		Welcome:           "👋 Добро пожаловать!\n\nВведите свои данные для регистрации.",
		AskFullName:       "👤 Введите имя и фамилию:\n\n<i>Пример: Анвар Каримов</i>",
		InvalidFullName:   "❌ Введите полное имя и фамилию.\n\n<i>Пример: Анвар Каримов</i>",
		AskLocation:       "📍 Введите область, район и школу:\n\n<i>Пример: Ташкент, Юнусабад, школа 56</i>",
		InvalidLocation:   "❌ Неверный формат.\n\nВведите в формате:\n<i>Область, Район, Школа</i>\n\nПример: Ташкент, Юнусабад, школа 56",
		AskGrade:          "🎓 В каком классе вы учитесь?\n\n<i>Введите число от 1 до 11</i>",
		InvalidGrade:      "❌ Неверный номер класса.\n\n<i>Введите число от 1 до 11</i>",
		AskPhone:          "📱 Введите номер телефона или нажмите кнопку ниже:\n\n<i>Пример: 998901234567</i>",
		AskOTP:            "🔐 Введите код подтверждения, отправленный на ваш телефон:",
		InvalidPhone:      "❌ Неверный формат номера.\n\n<i>Пример: 998901234567</i>",
		InvalidOTP:        "❌ Неверный код подтверждения. Попробуйте еще раз.",
		OTPSent:           "✅ Код подтверждения отправлен: <b>%s</b>",
		RegistrationDone:  "🎉 Поздравляем! Вы успешно зарегистрировались.\n\n👤 Имя: <b>%s</b>\n👤 Фамилия: <b>%s</b>\n🏙 Область: <b>%s</b>\n🏘 Район: <b>%s</b>\n🏫 Школа: <b>%s</b>\n🎓 Класс: <b>%d</b>\n📱 Телефон: <b>%s</b>",
		MainMenu:          "Ассалому алайкум! Добро пожаловать в наш бот.\n\nЕсли у вас есть аккаунт на платформе онлайн олимпиад <a href=\"https://khiso.uz\">khiso.uz</a>, нажмите кнопку \"Войти в аккаунт\".\nЕсли вы не зарегистрированы на платформе <a href=\"https://khiso.uz\">khiso.uz</a>, зарегистрируйтесь через кнопку \"Создать аккаунт\".\n\n<b>Внимание!</b> После создания аккаунта вы можете перейти в раздел Олимпиады и зарегистрироваться.",
		AlreadyRegistered: "✅ Вы уже зарегистрированы.",
		Error:             "❌ Произошла ошибка. Попробуйте еще раз.",
		ResendOTP:         "🔄 Отправить код повторно",
		NotRegistered:     "❌ Вы еще не зарегистрированы. Нажмите /start.",
		BtnLogin:          "🔑 Войти в аккаунт",
		BtnRegister:       "📝 Создать аккаунт",
		BtnShareContact:   "📱 Поделиться номером",
		MustSubscribe:     "📢 Для использования бота подпишитесь на следующие каналы:",
		BtnCheckSub:       "✅ Проверить подписку",
		SubscribeSuccess:  "✅ Спасибо! Теперь вы можете использовать бота.",
	},
	"en": {
		Welcome:           "👋 Welcome!\n\nPlease enter your information to register.",
		AskFullName:       "👤 Enter your first and last name:\n\n<i>Example: John Smith</i>",
		InvalidFullName:   "❌ Please enter your full name.\n\n<i>Example: John Smith</i>",
		AskLocation:       "📍 Enter your region, district and school:\n\n<i>Example: Tashkent, Yunusabad, School 56</i>",
		InvalidLocation:   "❌ Invalid format.\n\nPlease enter in format:\n<i>Region, District, School</i>\n\nExample: Tashkent, Yunusabad, School 56",
		AskGrade:          "🎓 What grade are you in?\n\n<i>Enter a number from 1 to 11</i>",
		InvalidGrade:      "❌ Invalid grade number.\n\n<i>Enter a number from 1 to 11</i>",
		AskPhone:          "📱 Enter your phone number or tap the button below:\n\n<i>Example: 998901234567</i>",
		AskOTP:            "🔐 Enter the verification code sent to your phone:",
		InvalidPhone:      "❌ Invalid phone number format.\n\n<i>Example: 998901234567</i>",
		InvalidOTP:        "❌ Invalid verification code. Please try again.",
		OTPSent:           "✅ Verification code sent: <b>%s</b>",
		RegistrationDone:  "🎉 Congratulations! You have successfully registered.\n\n👤 First Name: <b>%s</b>\n👤 Last Name: <b>%s</b>\n🏙 Region: <b>%s</b>\n🏘 District: <b>%s</b>\n🏫 School: <b>%s</b>\n🎓 Grade: <b>%d</b>\n📱 Phone: <b>%s</b>",
		MainMenu:          "Assalomu alaykum! Welcome to our bot.\n\nIf you have an account on the <a href=\"https://khiso.uz\">khiso.uz</a> online olympiad platform, click \"Login to account\".\nIf you are not registered on <a href=\"https://khiso.uz\">khiso.uz</a>, register via \"Create account\" button.\n\n<b>Attention!</b> After creating an account, you can go to the Olympiads section and register.",
		AlreadyRegistered: "✅ You are already registered.",
		Error:             "❌ An error occurred. Please try again.",
		ResendOTP:         "🔄 Resend code",
		NotRegistered:     "❌ You are not registered yet. Press /start.",
		BtnLogin:          "🔑 Login to account",
		BtnRegister:       "📝 Create account",
		BtnShareContact:   "📱 Share phone number",
		MustSubscribe:     "📢 To use the bot, please subscribe to the following channels:",
		BtnCheckSub:       "✅ Check subscription",
		SubscribeSuccess:  "✅ Thank you! You can now use the bot.",
	},
}

func Get(lang string) Messages {
	if m, ok := messages[lang]; ok {
		return m
	}
	return messages["uz"]
}

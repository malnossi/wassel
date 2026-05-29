// Syrian City Names Translation Mapping from cities.csv
export const cityNamesAR = {
  "Abu Kamal": "البوكمال",
  "Afrin": "عفرين",
  "Al-Bab": "الباب",
  "Al-Haffah": "الحفة",
  "Al-Hirak": "الحراك",
  "Al-Malikiyah": "المالكية",
  "Al-Mukharram": "المخرم",
  "Al-Qusayr": "القصير",
  "Al-Qutayfah": "القطيفة",
  "Al-Rastan": "الرستن",
  "Al-Safira": "السفيرة",
  "Al-Sanamayn": "الصنمين",
  "Al-Shaykh Badr": "الشيخ بدر",
  "Al-Suqaylabiyah": "السقيلبية",
  "Al-Tall": "التل",
  "Aleppo": "حلب",
  "An-Nabk": "النبك",
  "Ariha": "أريحا",
  "Ash-Shaykh Miskin": "الشيخ مسكين",
  "Atarib": "الأتارب",
  "Ayn al-Arab": "عين العرب",
  "Azaz": "أعزاز",
  "Baniyas": "بانياس",
  "Damascus": "دمشق",
  "Daraa": "درعا",
  "Darayya": "داريا",
  "Dayr Hafir": "دير حافر",
  "Deir ez-Zor": "دير الزور",
  "Douma": "دوما",
  "Duraykish": "دريكيش",
  "Fiq": "فيق",
  "Hama": "حماة",
  "Harem": "حارم",
  "Hasakah": "الحسكة",
  "Homs": "حمص",
  "Idlib": "ادلب",
  "Izraa": "ازرع",
  "Jableh": "جبلة",
  "Jarabulus": "جرابلس",
  "Jasim": "جاسم",
  "Jisr al-Shughur": "جسر الشغور",
  "Latakia": "اللاذقية",
  "Maarat al-Numan": "معرة النعمان",
  "Manbij": "منبج",
  "Masyaf": "مصياف",
  "Mayadin": "الميادين",
  "Mhardeh": "محردة",
  "Nawa": "نوى",
  "Palmyra": "تدمر",
  "Qamishli": "القامشلي",
  "Qardaha": "القرداحة",
  "Qatana": "قطنا",
  "Qudsaya": "قدسيا",
  "Quneitra": "القنيطرة",
  "Raqqa": "الرقة",
  "Ras al-Ayn": "رأس العين",
  "Safita": "صافيتا",
  "Salamiyah": "سلمية",
  "Salkhad": "صلخد",
  "Salqin": "سلقين",
  "Suwayda": "السويداء",
  "Tabqa": "الثورة",
  "Taldou": "تلدو",
  "Talkalakh": "تلكلخ",
  "Tartus": "طرطوس",
  "Tell Abyad": "تل أبيض",
  "Yabroud": "يبرود",
  "Zabadani": "الزبداني"
};

/**
 * Translates a given English city name or device name to Arabic if recognized.
 * Handles "(You)" suffix gracefully.
 */
export function translateCityName(name) {
  if (!name) return "";
  
  let cleanName = name.trim();
  let suffix = "";
  
  if (cleanName.endsWith(" (You)")) {
    cleanName = cleanName.substring(0, cleanName.length - 6).trim();
    suffix = " (أنت)";
  } else if (cleanName.endsWith(" (أنت)")) {
    cleanName = cleanName.substring(0, cleanName.length - 6).trim();
    suffix = " (أنت)";
  }
  
  const arName = cityNamesAR[cleanName];
  return arName ? arName + suffix : name;
}

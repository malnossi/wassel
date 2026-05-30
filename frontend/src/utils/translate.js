// Syrian City Names Translation Mapping from cities.csv
export const cityNamesAR = {
  "AbuKamal": "البوكمال",
  "Afrin": "عفرين",
  "AlBab": "الباب",
  "AlHaffah": "الحفة",
  "AlHirak": "الحراك",
  "AlMalikiyah": "المالكية",
  "AlMukharram": "المخرم",
  "AlQusayr": "القصير",
  "AlQutayfah": "القطيفة",
  "AlRastan": "الرستن",
  "AlSafira": "السفيرة",
  "AlSanamayn": "الصنمين",
  "AlShaykhBadr": "الشيخ بدر",
  "AlSuqaylabiyah": "السقيلبية",
  "AlTall": "التل",
  "Aleppo": "حلب",
  "AnNabk": "النبك",
  "Ariha": "أريحا",
  "AshShaykhMiskin": "الشيخ مسكين",
  "Atarib": "الأتارب",
  "AynAlArab": "عين العرب",
  "Azaz": "أعزاز",
  "Baniyas": "بانياس",
  "Damascus": "دمشق",
  "Daraa": "درعا",
  "Darayya": "داريا",
  "DayrHafir": "دير حافر",
  "DeirEzZor": "دير الزور",
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
  "JisrAlShughur": "جسر الشغور",
  "Latakia": "اللاذقية",
  "MaaratAlNuman": "معرة النعمان",
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
  "RasAlAyn": "رأس العين",
  "Safita": "صافيتا",
  "Salamiyah": "سلمية",
  "Salkhad": "صلخد",
  "Salqin": "سلقين",
  "Suwayda": "السويداء",
  "Tabqa": "الثورة",
  "Taldou": "تلدو",
  "Talkalakh": "تلكلخ",
  "Tartus": "طرطوس",
  "TellAbyad": "تل أبيض",
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

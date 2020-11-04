package tr

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestCheckTr(t *testing.T) {
  configPath := "./etc.test/tr_test"
  res := LoadLangs("./etc.test/langs1111.yaml")
  assert.Equal(t, false, res)

  res = LoadLangs("./etc.test/langs_bad.yaml")
  assert.Equal(t, false, res)
  
  res = LoadLangs("./etc.test/langs.yaml")
  assert.Equal(t, true, res)
  assert.Equal(t, 3, len(mapLangs))
  
  tr_json_need := "{\"ar_EG\":\"عربي\",\"en_US\":\"English\",\"ru_RU\":\"Русский\"}"
  tr_json := string(JSON())
  assert.Equal(t, tr_json_need, tr_json)

  LoadTrs(configPath + "/tr1")
  SetDef("Привет!")
  tr_str_need := "مرحبًا!"
  tr_str := Tr("ar_EG", "Привет!")
  assert.Equal(t, tr_str_need, tr_str)

  tr_str_need = "--TEXT NOT FOUND--"
  tr_str = Tr("ar_EG", "Привет :)")
  assert.Equal(t, tr_str_need, tr_str)

  tr_str_need = "--TEXT NOT FOUND--"
  tr_str = Tr("en_US", "Привет :)")
  assert.Equal(t, tr_str_need, tr_str)
  
  SaveNew(configPath + "/tr1")

  assert.Equal(t, 4, Count())
  assert.Equal(t, 3, LangCount())
  
  assert.Equal(t, "Русский", LangName("ru_RU"))
  assert.Equal(t, "English", LangName("en_US"))


  lang, ok := GetLocale()
  assert.Equal(t, true, ok)
  assert.Equal(t, "ru_RU", lang)

  lang_need := map[string]map[string]string(map[string]map[string]string{"ar_EG":map[string]string{"code":"ar_EG", "display_name":"عربي"}, "en_US":map[string]string{"code":"en_US", "display_name":"English"}, "ru_RU":map[string]string{"code":"ru_RU", "display_name":"Русский"}})
  langs := GetList()
  assert.Equal(t, &lang_need, langs)
}

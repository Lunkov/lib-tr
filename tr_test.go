package tr

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestCheckTr(t *testing.T) {
  tr := New()
  
  configPath := "./etc.test/tr_test"
  res := tr.LoadLangs("./etc.test/langs1111.yaml")
  assert.Equal(t, false, res)

  res = tr.LoadLangs("./etc.test/langs_bad.yaml")
  assert.Equal(t, false, res)
  
  res = tr.LoadLangs("./etc.test/langs.yaml")
  assert.Equal(t, true, res)
  assert.Equal(t, 3, len(tr.mapLangs))
  
  tr_json_need := "{\"ar_EG\":\"عربي\",\"en_US\":\"English\",\"ru_RU\":\"Русский\"}"
  tr_json := string(tr.JSON())
  assert.Equal(t, tr_json_need, tr_json)

  tr.LoadTrs(configPath + "/tr1")
  tr.SetDef("Привет!")
  tr_str_need := "مرحبًا!"
  tr_str, ok := tr.Tr("ar_EG", "Привет!")
  assert.Equal(t, true, ok)
  assert.Equal(t, tr_str_need, tr_str)

  tr_str_need = "Привет :-)"
  tr_str, ok = tr.Tr("ar_EG", tr_str_need)
  assert.Equal(t, false, ok)
  assert.Equal(t, tr_str_need, tr_str)

  tr_str, _ = tr.Tr("en_US", tr_str_need)
  assert.Equal(t, false, ok)
  assert.Equal(t, tr_str_need, tr_str)
  
  tr.SaveNew(configPath + "/tr1")

  assert.Equal(t, 4, tr.Count())
  assert.Equal(t, 3, tr.LangCount())
  
  assert.Equal(t, "Русский", tr.LangName("ru_RU"))
  assert.Equal(t, "English", tr.LangName("en_US"))


  lang, ok := GetLocale()
  assert.Equal(t, true, ok)
  assert.Equal(t, "ru_RU", lang)

  lang_need := map[string]map[string]string(map[string]map[string]string{"ar_EG":map[string]string{"code":"ar_EG", "display_name":"عربي"}, "en_US":map[string]string{"code":"en_US", "display_name":"English"}, "ru_RU":map[string]string{"code":"ru_RU", "display_name":"Русский"}})
  langs := tr.GetList()
  assert.Equal(t, &lang_need, langs)
}

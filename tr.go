package tr

import (
  "os"
  "os/exec"
  "sync"
  "path/filepath"
  
  "strings"

  "crypto/md5"
  "encoding/hex"

  "io/ioutil"
  "gopkg.in/yaml.v2"
  "encoding/json"
  "github.com/golang/glog"
)


type Texts   map[string]string

const langDef = "default"

type Tr struct {
  trmu          sync.RWMutex
  mapNeedTrs    map[string]Texts
  mapTrs        map[string]Texts
  mapLangs      map[string]string
  jsonTr        []byte // is string
}

func New() (*Tr) {
  return &Tr{
      mapNeedTrs: make(map[string]Texts),
      mapTrs    : make(map[string]Texts),
      mapLangs  : make(map[string]string),
  }
}

func GetLocale() (string, bool) {
  // Check the LANG environment variable, common on UNIX.
  envlang, ok := os.LookupEnv("LANG")
  if ok {
    if glog.V(9) {
      glog.Infof("DBG: os.LookupEnv = #%v ", envlang)
    }
    res := strings.Split(envlang, ".")[0]
    if res != "" {
      return res, true
    }
  }

  // Exec powershell Get-Culture on Windows.
  cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
  output, err := cmd.Output()
  if err == nil {
    if glog.V(9) {
      glog.Infof("DBG: powershell(Get-Culture) = #%v ", envlang)
    }
    res := strings.Trim(string(output), "\r\n")
    if res != "" {
      return res, true
    }
  }

  return "", false
}

func (t *Tr) Count() int {
  return len(t.mapTrs)
}

func (t *Tr) GetList() *map[string]map[string]string {
  res := make(map[string]map[string]string)
  for key, item := range t.mapLangs {
    res[key] = make(map[string]string)
    res[key]["code"] = key
    res[key]["display_name"] = item
  }
  return &res
}

func (t *Tr) LangCount() int {
  return len(t.mapLangs)
}

func (t *Tr) LangName(lang string) string {
  lang_name, ok := t.mapLangs[lang]
  if ok {
    return lang_name
  }
  return lang
}

func LangDefault() string {
  return langDef
}

func (t *Tr) SetDef(text string) {
  t.trmu.Lock()
  if t.mapTrs[langDef] == nil {
    t.mapTrs[langDef] = make(Texts)
  }
  t.mapTrs[langDef][getMD5Hash(text)] = text
  t.trmu.Unlock()
}

func (t *Tr) Tr(lang , text string) (string, bool) {
  key := getMD5Hash(text)
  trText, ok := t.mapTrs[lang][key]
  if ok {
    return trText, true
  }
  trText, ok = t.mapTrs[langDef][key]
  if !ok {
    trText = text
  }
  t.trmu.Lock()
  if t.mapNeedTrs[lang] == nil {
    t.mapNeedTrs[lang] = make(Texts)
  }
  t.mapNeedTrs[lang][key] = text
  t.trmu.Unlock()
  return trText, false
}

func getMD5Hash(text string) string {
  hasher := md5.New()
  hasher.Write([]byte(text))
  return hex.EncodeToString(hasher.Sum(nil))
}

func (t *Tr) JSON() []byte {
  return t.jsonTr
}

func (t *Tr) LoadLangs(filename string) bool {
  yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    glog.Errorf("ERR: ReadFile.yamlFile(%s)  #%v ", filename, err)
    return false
  } else {
    err = yaml.Unmarshal(yamlFile, &t.mapLangs)
    if err != nil {
      glog.Errorf("ERR: yamlFile(%s): YAML: %v", filename, err)
      return false
    }
  }
  t.jsonTr, _ = json.Marshal(t.mapLangs)
  t.mapNeedTrs[langDef] = make(Texts)
  t.mapTrs[langDef] = make(Texts)
  for i, _ := range t.mapLangs {
    t.mapTrs[i] = make(Texts)
    t.mapNeedTrs[i] = make(Texts)
  }
  return true
}

func (t *Tr) LoadTrs(scanPath string) bool {
  for lang_code, _ := range t.mapLangs {
    filepath.Walk(scanPath + "/" + lang_code, func(filename string, f os.FileInfo, err error) error {
      if f != nil && f.IsDir() == false && filepath.Ext(filename) == ".yaml" {
        t.loadTrs(lang_code, filename)
      }
      return nil
    })
  }
  return true
}

func (t *Tr) loadTrs(lang, filename string) bool {
  yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    glog.Errorf("ERR: ReadFile.yamlFile(%s)  #%v ", filename, err)
    return false
  } else {
    var mapTmp Texts
    err = yaml.Unmarshal(yamlFile, &mapTmp)
    if err != nil {
      glog.Errorf("ERR: yamlFile(%s): YAML: %v", filename, err)
      return false
    }
    for key, text := range mapTmp {
      t.mapTrs[lang][key] = text
    }
  }
  return true
}

func (t *Tr) SaveNew(scanPath string) bool {
  for lang_code, _ := range t.mapLangs {
    mapTmp := t.mapNeedTrs[lang_code]
    for key, text := range t.mapTrs[langDef] {
      _, ok := t.mapTrs[lang_code][key]
      if !ok {
        mapTmp[key] = text
        if glog.V(9) {
          glog.Infof("DBG: NEED TR(%s) = #%s ", lang_code, text)
        }
      }
    }
    if len(mapTmp) > 0 {
      d, err := yaml.Marshal(&mapTmp)
      if err != nil {
        glog.Errorf("ERR: TR: Marshal, err = %v", err)
      } else {
        folderPath := scanPath + "/" + lang_code
        filename := folderPath + "/tr_new.!yaml"
        os.MkdirAll(folderPath, os.ModePerm)
        err = ioutil.WriteFile(filename, d, 0644)
        if err != nil {
          glog.Errorf("ERR: SAVE TRs: (%s) err=%s \n", filename, err)
        } else {
          glog.Infof("LOG: TR: Need translate (count=%d), file = %s", len(mapTmp), filename)
        }
      }
    }
  }
  return true
}

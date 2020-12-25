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
var trmu sync.RWMutex
var mapNeedTrs   = make(map[string]Texts)
var mapTrs       = make(map[string]Texts)
var mapLangs     = make(map[string]string)
var jsonTr []byte // is string

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

func Count() int {
  return len(mapTrs)
}

func GetList() *map[string]map[string]string {
  res := make(map[string]map[string]string)
  for key, item := range mapLangs {
    res[key] = make(map[string]string)
    res[key]["code"] = key
    res[key]["display_name"] = item
  }
  return &res
}

func LangCount() int {
  return len(mapLangs)
}

func LangName(lang string) string {
  lang_name, ok := mapLangs[lang]
  if ok {
    return lang_name
  }
  return lang
}

func LangDefault() string {
  return langDef
}

func SetDef(text string) {
  trmu.Lock()
  if mapTrs[langDef] == nil {
    mapTrs[langDef] = make(Texts)
  }
  mapTrs[langDef][getMD5Hash(text)] = text
  trmu.Unlock()
}

func Tr(lang , text string) (string, bool) {
  key := getMD5Hash(text)
  tr, ok := mapTrs[lang][key]
  if ok {
    return tr, true
  }
  tr, ok = mapTrs[langDef][key]
  if ok {
    return tr, false
  }
  trmu.Lock()
  if mapNeedTrs[lang] == nil {
    mapNeedTrs[lang] = make(Texts)
  }
  mapNeedTrs[lang][key] = text
  trmu.Unlock()
  return text, false
}

func getMD5Hash(text string) string {
  hasher := md5.New()
  hasher.Write([]byte(text))
  return hex.EncodeToString(hasher.Sum(nil))
}

func JSON() []byte {
  return jsonTr
}

func LoadLangs(filename string) bool {
  yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    glog.Errorf("ERR: ReadFile.yamlFile(%s)  #%v ", filename, err)
    return false
  } else {
    err = yaml.Unmarshal(yamlFile, &mapLangs)
    if err != nil {
      glog.Errorf("ERR: yamlFile(%s): YAML: %v", filename, err)
      return false
    }
  }
  jsonTr, _ = json.Marshal(mapLangs)
  mapNeedTrs[langDef] = make(Texts)
  mapTrs[langDef] = make(Texts)
  for i, _ := range mapLangs {
    mapTrs[i] = make(Texts)
    mapNeedTrs[i] = make(Texts)
  }
  return true
}

func LoadTrs(scanPath string) bool {
  for lang_code, _ := range mapLangs {
    filepath.Walk(scanPath + "/" + lang_code, func(filename string, f os.FileInfo, err error) error {
      if f != nil && f.IsDir() == false && filepath.Ext(filename) == ".yaml" {
        loadTrs(lang_code, filename)
      }
      return nil
    })
  }
  return true
}

func loadTrs(lang, filename string) bool {
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
      mapTrs[lang][key] = text
    }
  }
  return true
}

func SaveNew(scanPath string) bool {
  for lang_code, _ := range mapLangs {
    mapTmp := mapNeedTrs[lang_code]
    for key, text := range mapTrs[langDef] {
      _, ok := mapTrs[lang_code][key]
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

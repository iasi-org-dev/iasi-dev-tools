// Package versioning implements SemVer 2.0.0 and temporary legacy tag compatibility.
package versioning

import (
 "regexp"
 "strconv"
 "strings"
)

var pattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-((?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`)

type Version struct { Major, Minor, Patch uint64; Pre string }

func Parse(s string) (Version, bool) {
 m:=pattern.FindStringSubmatch(s)
 if m==nil { return Version{},false }
 a,e1:=strconv.ParseUint(m[1],10,64);b,e2:=strconv.ParseUint(m[2],10,64);c,e3:=strconv.ParseUint(m[3],10,64)
 if e1!=nil||e2!=nil||e3!=nil{return Version{},false}
 return Version{a,b,c,m[4]},true
}

// ParseLegacy is used only when inspecting previously published Git tags or the
// organization's VERSION during its one-time migration.
func ParseLegacy(s string)(Version,bool){
 if v,ok:=Parse(s);ok{return v,true}
 if strings.HasPrefix(s,"v"){return Parse(strings.TrimPrefix(s,"v"))}
 return Version{},false
}

func Compare(a,b Version) int {
 for i,x:=range []uint64{a.Major,a.Minor,a.Patch}{y:=[]uint64{b.Major,b.Minor,b.Patch}[i];if x<y{return -1};if x>y{return 1}}
 if a.Pre==b.Pre{return 0};if a.Pre==""{return 1};if b.Pre==""{return -1}
 aa,bb:=strings.Split(a.Pre,"."),strings.Split(b.Pre,".")
 for i:=0;i<len(aa)&&i<len(bb);i++{
  if aa[i]==bb[i]{continue}
  an,ae:=strconv.ParseUint(aa[i],10,64);bn,be:=strconv.ParseUint(bb[i],10,64)
  if ae==nil&&be==nil {if an<bn{return -1};return 1}
  if ae==nil{return -1};if be==nil{return 1}
  if aa[i]<bb[i]{return -1};return 1
 }
 if len(aa)<len(bb){return -1};if len(aa)>len(bb){return 1};return 0
}

// TagCandidates prefers the canonical Git tag, falling back to legacy v tags.
func TagCandidates(version string) []string {return []string{version,"v"+version}}

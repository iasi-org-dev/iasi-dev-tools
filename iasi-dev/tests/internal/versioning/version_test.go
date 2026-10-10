package versioning
import "testing"
func TestSemVer(t *testing.T){
 for _,s:=range []string{"0.6.0","1.0.0+build.154","1.0.0-rc.1","1.2.3-alpha.2+build.4"}{if _,ok:=Parse(s);!ok{t.Fatalf("rejected %s",s)}}
 for _,s:=range []string{"v0.6.0","0.6","0.6.0.1","01.2.3","1.2.3-alpha.01"}{if _,ok:=Parse(s);ok{t.Fatalf("accepted %s",s)}}
 a,_:=Parse("1.0.0-rc.1");b,_:=Parse("1.0.0");if Compare(a,b)>=0{t.Fatal("prerelease ordering")}
 a,_=Parse("1.0.0+build.1");b,_=Parse("1.0.0+build.2");if Compare(a,b)!=0{t.Fatal("build must not affect precedence")}
 if _,ok:=ParseLegacy("v0.5.6");!ok{t.Fatal("legacy tag rejected")}
}

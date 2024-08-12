package demo

import (
	"fmt"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	s := "戴胄和萧瑀是多年好友，自然不会因为萧瑀这玩笑话生气。\n\n\n但他也知道，看萧瑀这样子，倘若林枫不死，是绝对不会撒手的。\n\n\n也是，这样一个人才，若是在自己手里，谁要是抢走了，那可不比夺妻之恨要轻。\n\n\n“鬼影之谜算是彻底解开了，不过……究竟是谁布置出这样一个精巧绝妙的机关的？”"

	sArr := strings.Split(s, "\n")

	for k, v := range sArr {
		if len(v) == 0 {
			sArr[k] = "[remove@]"
		}
	}
	s = strings.Join(sArr, "\n")
	fmt.Println(s)
	s = strings.ReplaceAll(s, "[remove@]\n", "")
	s = strings.ReplaceAll(s, "[remove@]", "")
	fmt.Println(s)

}

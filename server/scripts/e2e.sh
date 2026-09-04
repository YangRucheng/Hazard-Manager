#!/usr/bin/env bash
# 端到端测试：真实启动后端（连 MySQL）后，通过 HTTP 模拟管理员核心操作全流程。
# 覆盖：鉴权、枚举、图片摘要去重、隐患默认值/联动/分类归属、统计、删除保护、软删除、错误路径。
# 依赖：curl、jq、（生成测试图片用）python3。
# 用法：E2E_BASE_URL=http://127.0.0.1:8090 ADMIN_PASSWORD=xxx bash server/scripts/e2e.sh
set -euo pipefail

BASE_URL="${E2E_BASE_URL:-http://127.0.0.1:8090}"
ADMIN_USERNAME="${ADMIN_USERNAME:-admin}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-admin123456}"

PASS=0
FAIL=0
FAILED_NAMES=()

# 断言：相等
assert_eq() {
  local name="$1" expected="$2" actual="$3"
  if [[ "$expected" == "$actual" ]]; then
    PASS=$((PASS + 1))
    echo "  ✔ $name"
  else
    FAIL=$((FAIL + 1))
    FAILED_NAMES+=("$name")
    echo "  ✘ $name（期望 $expected，实际 $actual）"
  fi
}

# 断言：HTTP 状态码
assert_status() {
  local name="$1" expected="$2" code="$3"
  if [[ "$expected" == "$code" ]]; then
    PASS=$((PASS + 1))
    echo "  ✔ $name"
  else
    FAIL=$((FAIL + 1))
    FAILED_NAMES+=("$name")
    echo "  ✘ $name（期望 HTTP $expected，实际 HTTP $code）"
  fi
}

# 请求辅助：输出 HTTP 状态码
http_code() { curl -s -o /dev/null -w '%{http_code}' "$@"; }

echo "== 0. 健康检查 =="
assert_status "healthz 200" 200 "$(http_code "$BASE_URL/healthz")"

echo "== 1. 鉴权 =="
assert_status "错误密码登录 401" 401 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"wrong-password\"}")"
assert_status "未带令牌访问列表 401" 401 "$(http_code "$BASE_URL/api/v1/hazards")"

TOKEN="$(curl -s -X POST "$BASE_URL/api/v1/auth/login" -H 'Content-Type: application/json' -d "{\"username\":\"$ADMIN_USERNAME\",\"password\":\"$ADMIN_PASSWORD\"}" | jq -r '.token')"
assert_eq "登录返回 token 非空" "true" "$([[ -n "$TOKEN" && "$TOKEN" != "null" ]] && echo true || echo false)"
AUTH=(-H "Authorization: Bearer $TOKEN")

ME="$(curl -s "$BASE_URL/api/v1/auth/me" "${AUTH[@]}")"
assert_eq "/auth/me 用户名" "$ADMIN_USERNAME" "$(echo "$ME" | jq -r '.username')"
assert_eq "/auth/me 用户类型" "admin" "$(echo "$ME" | jq -r '.userType')"

echo "== 2. 枚举 =="
UNITS="$(curl -s "$BASE_URL/api/v1/units" "${AUTH[@]}")"
assert_eq "单位列表非空" "true" "$([[ "$(echo "$UNITS" | jq 'length')" > 0 ]] && echo true || echo false)"
UNIT_ID="$(echo "$UNITS" | jq -r '.[0].id')"
UNIT_NAME="$(echo "$UNITS" | jq -r '.[0].name')"
UNIT_PERSON="$(echo "$UNITS" | jq -r '.[0].person')"

TYPES="$(curl -s "$BASE_URL/api/v1/hazard-types" "${AUTH[@]}")"
assert_eq "类型组合数 ≥5" "true" "$([[ "$(echo "$TYPES" | jq 'length')" -ge 5 ]] && echo true || echo false)"
# 取 seed 中的「电气设备 / 线路老化」组合行作为隐患类型
TYPE_ID="$(echo "$TYPES" | jq -r '.[] | select(.major=="电气设备" and .minor=="线路老化") | .id' | head -1)"
TYPE_MAJOR="电气设备"
TYPE_MINOR="线路老化"
assert_eq "存在该组合" "true" "$([[ -n "$TYPE_ID" && "$TYPE_ID" != "null" ]] && echo true || echo false)"
# 枚举维护：新增（复用既有大类 + 新小类）、重复组合 422、全新大类、改名、删除未引用类型
NEWTYPE="$(curl -s -X POST "$BASE_URL/api/v1/hazard-types" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"major":"电气设备","minor":"E2E新小类"}')"
NEWTYPE_ID="$(echo "$NEWTYPE" | jq -r '.id')"
assert_eq "新增组合成功" "true" "$([[ -n "$NEWTYPE_ID" && "$NEWTYPE_ID" != "null" ]] && echo true || echo false)"
assert_eq "新增组合大类正确" "电气设备" "$(echo "$NEWTYPE" | jq -r '.major')"
assert_status "重复组合 422" 422 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/hazard-types" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"major":"电气设备","minor":"E2E新小类"}')"
BRANDNEW="$(curl -s -X POST "$BASE_URL/api/v1/hazard-types" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"major":"机械伤害","minor":"旋转部位无防护"}')"
BRANDNEW_ID="$(echo "$BRANDNEW" | jq -r '.id')"
assert_eq "支持全新大类" "机械伤害" "$(echo "$BRANDNEW" | jq -r '.major')"
assert_status "更新组合 200" 200 "$(curl -s -o /dev/null -w '%{http_code}' -X PUT "$BASE_URL/api/v1/hazard-types/$BRANDNEW_ID" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"minor":"旋转部位无防护（改）"}')"
assert_status "删除未引用类型 204" 204 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/hazard-types/$BRANDNEW_ID" "${AUTH[@]}")"
# 缺大类/小类 422
assert_status "缺小类 422" 422 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/hazard-types" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"major":"电气设备"}')"

echo "== 3. 图片上传与摘要去重 =="
TMPDIR_E2E="$(mktemp -d)"
python3 - "$TMPDIR_E2E" <<'PY'
import base64, sys, os, zlib, struct
d = sys.argv[1]
png = base64.b64decode('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==')
open(os.path.join(d, 'a.png'), 'wb').write(png)
open(os.path.join(d, 'b.png'), 'wb').write(png)  # 与 a 相同内容
def chunk(t, data):
    c = struct.pack('>I', len(data)) + t + data
    return c + struct.pack('>I', zlib.crc32(t + data) & 0xffffffff)
ihdr = struct.pack('>IIBBBBB', 2, 2, 8, 2, 0, 0, 0)
raw = b''.join(b'\x00' + bytes([r, g, b]) * 2 for r, g, b in [(255, 0, 0), (0, 0, 255)])
png2 = b'\x89PNG\r\n\x1a\n' + chunk(b'IHDR', ihdr) + chunk(b'IDAT', zlib.compress(raw)) + chunk(b'IEND', b'')
open(os.path.join(d, 'c.png'), 'wb').write(png2)
PY

UP_A="$(curl -s -X POST "$BASE_URL/api/v1/images" "${AUTH[@]}" -F "file=@$TMPDIR_E2E/a.png")"
IMG_ID="$(echo "$UP_A" | jq -r '.id')"
assert_eq "上传 a 非重复" "false" "$(echo "$UP_A" | jq -r '.duplicate')"
UP_B="$(curl -s -X POST "$BASE_URL/api/v1/images" "${AUTH[@]}" -F "file=@$TMPDIR_E2E/b.png")"
assert_eq "上传 b 命中去重" "true" "$(echo "$UP_B" | jq -r '.duplicate')"
assert_eq "同摘要返回同一 uuid" "$IMG_ID" "$(echo "$UP_B" | jq -r '.id')"
IMG_ID2="$(curl -s -X POST "$BASE_URL/api/v1/images" "${AUTH[@]}" -F "file=@$TMPDIR_E2E/c.png" | jq -r '.id')"
assert_eq "不同内容返回新 uuid" "true" "$([[ "$IMG_ID2" != "$IMG_ID" ]] && echo true || echo false)"
assert_status "获取原图 200" 200 "$(http_code "$BASE_URL/api/v1/images/$IMG_ID" "${AUTH[@]}")"
assert_status "获取缩略图 200" 200 "$(http_code "$BASE_URL/api/v1/images/$IMG_ID/thumbnail" "${AUTH[@]}")"
assert_status "不存在的图片 404" 404 "$(http_code "$BASE_URL/api/v1/images/00000000000000000000000000000000" "${AUTH[@]}")"
assert_status "非法文件类型 400" 400 "$(printf 'not an image' > "$TMPDIR_E2E/bad.txt"; curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/images" "${AUTH[@]}" -F "file=@$TMPDIR_E2E/bad.txt")"

echo "== 4. 新增隐患（默认值与联动） =="
# 缺必填 422
assert_status "缺责任单位 422" 422 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/hazards" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"description":"缺单位测试","typeId":'"$TYPE_ID"'}')"
# 类型不存在 422
assert_status "类型不存在 422" 422 "$(curl -s -o /dev/null -w '%{http_code}' -X POST "$BASE_URL/api/v1/hazards" "${AUTH[@]}" -H 'Content-Type: application/json' -d "{\"description\":\"类型不存在\",\"unitId\":$UNIT_ID,\"typeId\":99999}")"

TODAY="$(date +%F)"
DUE_EXPECTED="$(date -d "$TODAY +7 days" +%F 2>/dev/null || date -v+7d +%F)"
CREATE_BODY="{\"description\":\"E2E 配电箱线路老化，存在漏电风险\",\"suggestion\":\"更换老化线路\",\"unitId\":$UNIT_ID,\"typeId\":$TYPE_ID,\"beforeImageIds\":[\"$IMG_ID\",\"$IMG_ID2\"],\"level\":\"重大隐患\",\"remark\":\"E2E 备注\"}"
CREATED="$(curl -s -X POST "$BASE_URL/api/v1/hazards" "${AUTH[@]}" -H 'Content-Type: application/json' -d "$CREATE_BODY")"
HAZARD_ID="$(echo "$CREATED" | jq -r '.id')"
assert_eq "检查区域默认华星现场" "华星现场" "$(echo "$CREATED" | jq -r '.inspectionArea')"
assert_eq "检查日期默认今天" "$TODAY" "$(echo "$CREATED" | jq -r '.inspectionDate')"
assert_eq "检查人员默认电气自查" "电气自查" "$(echo "$CREATED" | jq -r '.inspector')"
assert_eq "检查日期+7 联动" "$DUE_EXPECTED" "$(echo "$CREATED" | jq -r '.dueDate')"
assert_eq "复查人员默认检查人员" "电气自查" "$(echo "$CREATED" | jq -r '.recheckPerson')"
assert_eq "责任人由单位联动" "$UNIT_PERSON" "$(echo "$CREATED" | jq -r '.person')"
assert_eq "状态默认待整改" "待整改" "$(echo "$CREATED" | jq -r '.status')"
assert_eq "等级指定重大隐患" "重大隐患" "$(echo "$CREATED" | jq -r '.level')"
assert_eq "整改前图片数量 2" "2" "$(echo "$CREATED" | jq '.beforeImageIds | length')"

echo "== 5. 列表 / 详情（JOIN 名称） =="
assert_eq "详情单位名称非空" "$UNIT_NAME" "$(curl -s "$BASE_URL/api/v1/hazards/$HAZARD_ID" "${AUTH[@]}" | jq -r '.unitName')"
DETAIL_TYPE_MAJOR="$(curl -s "$BASE_URL/api/v1/hazards/$HAZARD_ID" "${AUTH[@]}" | jq -r '.typeMajor')"
DETAIL_TYPE_MINOR="$(curl -s "$BASE_URL/api/v1/hazards/$HAZARD_ID" "${AUTH[@]}" | jq -r '.typeMinor')"
assert_eq "详情类型大类正确" "$TYPE_MAJOR" "$DETAIL_TYPE_MAJOR"
assert_eq "详情类型小类正确" "$TYPE_MINOR" "$DETAIL_TYPE_MINOR"
LIST_TOTAL="$(curl -s "$BASE_URL/api/v1/hazards" "${AUTH[@]}" | jq -r '.pagination.total')"
assert_eq "列表包含新增记录" "true" "$([[ "$LIST_TOTAL" -ge 1 ]] && echo true || echo false)"
assert_status "详情 404（不存在）" 404 "$(http_code "$BASE_URL/api/v1/hazards/999999" "${AUTH[@]}")"

echo "== 6. 更新（状态流转 + 单位联动） =="
UPDATED="$(curl -s -X PUT "$BASE_URL/api/v1/hazards/$HAZARD_ID" "${AUTH[@]}" -H 'Content-Type: application/json' -d '{"status":"已整改","afterImageIds":["'"$IMG_ID2"'"],"recheckPerson":"E2E复核"}')"
assert_eq "状态更新为已整改" "已整改" "$(echo "$UPDATED" | jq -r '.status')"
assert_eq "整改后图片数量 1" "1" "$(echo "$UPDATED" | jq '.afterImageIds | length')"
assert_eq "复查人员更新" "E2E复核" "$(echo "$UPDATED" | jq -r '.recheckPerson')"

echo "== 7. 统计 =="
STATS="$(curl -s "$BASE_URL/api/v1/hazards/stats" "${AUTH[@]}")"
assert_eq "已整改数 ≥1" "true" "$([[ "$(echo "$STATS" | jq -r '.done')" -ge 1 ]] && echo true || echo false)"

echo "== 8. 删除保护 =="
assert_status "删除被引用单位 409" 409 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/units/$UNIT_ID" "${AUTH[@]}")"
assert_status "删除被引用类型 409" 409 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/hazard-types/$TYPE_ID" "${AUTH[@]}")"
assert_status "删除未引用类型 204" 204 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/hazard-types/$NEWTYPE_ID" "${AUTH[@]}")"
assert_status "删除不存在单位 404" 404 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/units/999999" "${AUTH[@]}")"

echo "== 9. 软删除 =="
assert_status "删除隐患 204" 204 "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE "$BASE_URL/api/v1/hazards/$HAZARD_ID" "${AUTH[@]}")"
AFTER_TOTAL="$(curl -s "$BASE_URL/api/v1/hazards" "${AUTH[@]}" | jq '.pagination.total')"
TOTAL_AFTER_DELETE="$((LIST_TOTAL - 1))"
assert_eq "删除后列表总数 -1" "$TOTAL_AFTER_DELETE" "$AFTER_TOTAL"

rm -rf "$TMPDIR_E2E"

echo ""
echo "================ 结果汇总 ================"
echo "通过：$PASS  失败：$FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  printf '失败项：%s\n' "${FAILED_NAMES[@]}"
  exit 1
fi
echo "✅ E2E 全部通过"
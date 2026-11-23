"""Drive the exported app; only LINE identity and session establishment are fixtures."""
import shutil
import sys
from urllib.parse import parse_qs, urlparse
from playwright.sync_api import expect, sync_playwright

base, actor, organization = sys.argv[1:]
with sync_playwright() as playwright:
    browser = playwright.chromium.launch(executable_path=shutil.which("chromium"), args=["--no-sandbox"])
    page = browser.new_page()
    page.add_init_script("""window.liff = {init: async()=>{}, isLoggedIn:()=>true,
      isInClient:()=>false, getIDToken:()=>"fixture", getProfile:async()=>({displayName:"Tester"})};""")

    def route_request(route):
        path = urlparse(route.request.url)
        if not route.request.url.startswith(base + "/"):
            route.abort()
        elif path.path == "/api/v1/miniapp/config":
            route.fulfill(json={"liffId": "fixture"})
        elif path.path == "/api/v1/session":
            route.fulfill(json={"user": {"userId":actor,"displayName":"Tester"},
                                "csrfToken":"c"*43,"expiresAt":"2099-01-01T00:00:00Z"})
        else:
            route.continue_(headers={**route.request.headers, "X-Test-Actor":actor})

    page.route("**/*", route_request)
    button = lambda name: page.get_by_role("button", name=name, exact=True)
    field = page.get_by_label
    page.goto(base + "/miniapp/projects")
    button("新增專案").click()
    field("標題", exact=True).fill("Maintenance")
    button("建立專案").click()
    button("新增欄位").click()
    original = page.url
    field("欄位名稱").fill("Priority")
    field("選項（每行一項）").fill("High\nLow")
    button("新增檢視").click()
    field("檢視名稱").fill("Triage")
    button("儲存內容").click()
    expect(page.get_by_text("已更新專案。", exact=True)).to_be_visible()
    field("複製到另一個擁有者").select_option(organization)
    field("標題", exact=True).fill("Audit")
    button("建立專案").click()
    expect(field("欄位名稱")).to_have_value("Priority")
    copied = page.url
    assert parse_qs(urlparse(copied).query)["owner"] == [organization]
    assert parse_qs(urlparse(copied).query)["project"] != parse_qs(urlparse(original).query)["project"]
    field("欄位名稱").fill("Severity")
    button("儲存內容").click()
    expect(page.get_by_text("已更新專案。", exact=True)).to_be_visible()
    page.reload()
    for url, title, name in [(copied,"Audit","Severity"),(original,"Maintenance","Priority")]:
        page.goto(url)
        expect(field("標題",exact=True)).to_have_value(title)
        expect(field("欄位名稱")).to_have_value(name)
        expect(field("選項（每行一項）")).to_have_value("High\nLow")
        expect(field("檢視名稱")).to_have_value("Triage")
        expect(field("檢視欄位").locator("option:checked")).to_have_text(name)
    browser.close()

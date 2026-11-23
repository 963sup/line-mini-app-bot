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
    field("標題", exact=True).fill("Unsaved project")
    draft_url = page.url
    page.once("dialog", lambda dialog: dialog.dismiss())
    with page.expect_event("dialog"):
        page.evaluate("history.back()")
    expect(page).to_have_url(draft_url)
    expect(field("標題", exact=True)).to_have_value("Unsaved project")
    page.once("dialog", lambda dialog: dialog.accept())
    with page.expect_event("dialog"):
        page.evaluate("history.back()")
    expect(button("新增專案")).to_be_visible()
    page.evaluate("history.forward()")
    expect(page).to_have_url(draft_url)
    expect(field("標題", exact=True)).to_have_value("")
    field("標題", exact=True).fill("Maintenance")
    personal_projects = f"{base}/api/v1/owners/{actor}/projects"
    committed_ids = []

    def drop_create_response(route):
        assert route.request.method == "POST"
        response = route.fetch(headers={**route.request.headers, "X-Test-Actor": actor})
        assert response.ok, response.status
        committed_ids.append(response.json()["id"])
        route.abort("connectionreset")

    page.route(personal_projects, drop_create_response, times=1)
    button("建立專案").click()
    button("重試原建立請求").click()
    button("新增欄位").click()
    assert parse_qs(urlparse(page.url).query)["project"] == committed_ids
    listing = page.request.get(personal_projects, headers={"X-Test-Actor": actor})
    assert listing.ok, listing.status
    assert [project["id"] for project in listing.json()["items"]] == committed_ids
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

import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../navigation/home/View";
import { useRouter } from "next/router";
import { OrganizationPage } from "@/business/organization/OrganizationPage";
import { Enterprise } from "@/business/enterprise/Enterprise";
import { OrganizationTeams } from "@/business/team/OrganizationTeams";
import { useOrganizationSelection } from "@/business/organization/OrganizationSelection";
export default function Page() {
  const router = useRouter();
  const { query } = router;
  const { refresh } = useOrganizationSelection();
  const id = typeof query.organization === "string" ? query.organization : "";
  const enterprise =
    typeof query.enterprise === "string" ? query.enterprise : "";
  return (
    <Shell
      title={enterprise ? "企業" : id ? "組織" : "首頁"}
      workflow="home"
    >
      {enterprise ? (
        <Enterprise key={enterprise} id={enterprise} />
      ) : id ? (
        <OrganizationPage key={id} organizationId={id} onChanged={refresh} onLeave={() => { void router.replace("/miniapp"); }}>
          <OrganizationTeams organizationId={id} />
          <section><h2>儲存庫</h2><p>尚未支援組織儲存庫。</p></section>
        </OrganizationPage>
      ) : (
        <View />
      )}
    </Shell>
  );
}

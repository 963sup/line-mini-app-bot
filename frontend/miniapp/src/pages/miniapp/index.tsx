import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../navigation/home/View";
import { useRouter } from "next/router";
import { OrganizationTeams } from "@/business/team/OrganizationTeams";
import { Enterprise } from "@/business/enterprise/Enterprise";
export default function Page() {
  const { query } = useRouter();
  const id = typeof query.organization === "string" ? query.organization : "";
  const enterprise =
    typeof query.enterprise === "string" ? query.enterprise : "";
  return (
    <Shell
      title={enterprise ? "企業" : id ? "組織團隊" : "首頁"}
      workflow="home"
    >
      {enterprise ? (
        <Enterprise key={enterprise} id={enterprise} />
      ) : id ? (
        <OrganizationTeams key={id} organizationId={id} />
      ) : (
        <View />
      )}
    </Shell>
  );
}

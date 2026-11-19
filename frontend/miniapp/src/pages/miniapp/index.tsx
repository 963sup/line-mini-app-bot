import React from "react";
import { Shell } from "../../shell/Shell";
import { View } from "../../navigation/home/View";
import {useRouter} from "next/router";
import {OrganizationTeams} from "@/shell/OrganizationTeams";
export default function Page() {
  const { query } = useRouter();
  const id = typeof query.organization === "string" ? query.organization : "";
  return (
    <Shell title={id ? "組織團隊" : "首頁"} workflow="home">
      {id ? <OrganizationTeams key={id} organizationId={id} /> : <View />}
    </Shell>
  );
}

import { Shell } from "../../shell/Shell";
import { View } from "../../navigation/home/View";
import { useRouter } from "next/router";
import { OrganizationPage } from "@/business/organization/OrganizationPage";
import { useOrganizationSelection } from "@/business/organization/OrganizationSelection";
import { OrganizationTeams } from "@/business/team/OrganizationTeams";
import { Enterprise } from "@/business/enterprise/Enterprise";
export default function Page() {
  const router = useRouter();
  const { query } = router;
  const { context, error, refresh } = useOrganizationSelection();
  const id = typeof query.organization === "string" ? query.organization : "";
  const enterprise =
    typeof query.enterprise === "string" ? query.enterprise : "";
  return (
    <Shell
      title={enterprise ? "企業" : id ? "組織" : "首頁"}
      workflow="home"
    >
      {enterprise ? (
        <Enterprise
          key={enterprise}
          id={enterprise}
          canCreate={context?.canCreateEnterprise}
          eligibilityError={error}
          onSaved={enterpriseId => window.location.assign("/miniapp?" + new URLSearchParams({ enterprise: enterpriseId }))}
        />
      ) : id ? (
        <OrganizationPage key={id} organizationId={id} onChanged={refresh} onLeave={() => { void router.replace("/miniapp"); }}>
          <OrganizationTeams organizationId={id} />
        </OrganizationPage>
      ) : (
        <View />
      )}
    </Shell>
  );
}

<template>
  <BBTable
    :column-list="columnList"
    :data-source="identityProviderList"
    :show-header="true"
    :left-bordered="false"
    :right-bordered="false"
    @click-row="clickIdentityProvider"
  >
    <template #body="{ rowData: identityProvider }">
      <BBTableCell class="w-48 table-cell pl-4">
        {{ identityProvider.title }}
      </BBTableCell>
      <BBTableCell class="w-48 table-cell">
        {{ identityProvider.name }}
      </BBTableCell>
      <BBTableCell class="w-48 table-cell">
        {{ identityProvider.domain }}
      </BBTableCell>
    </template>
  </BBTable>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { BBTable, BBTableCell } from "@/bbkit";
import { WORKSPACE_ROUTE_SSO_DETAIL } from "@/router/dashboard/workspaceRoutes";
import { getSSOId } from "@/store/modules/v1/common";
import type { IdentityProvider } from "@/types/proto/v1/idp_service";

const props = defineProps<{
  identityProviderList: IdentityProvider[];
}>();

const router = useRouter();

const { t } = useI18n();

const columnList = computed(() => [
  {
    title: t("common.name"),
  },
  {
    title: t("settings.sso.form.resource-id"),
  },
  {
    title: t("settings.sso.form.domain"),
  },
]);

const clickIdentityProvider = function (section: number, row: number) {
  const identityProvider = props.identityProviderList[row];
  router.push({
    name: WORKSPACE_ROUTE_SSO_DETAIL,
    params: {
      ssoId: getSSOId(identityProvider.name),
    },
  });
};
</script>

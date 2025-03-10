<template>
  <PagedTable
    ref="projectPagedTable"
    :session-key="sessionKey"
    :fetch-list="fetchProjects"
  >
    <template #table="{ list, loading }">
      <ProjectV1Table v-bind="$attrs" :loading="loading" :project-list="list" />
    </template>
  </PagedTable>
</template>

<script lang="tsx" setup>
import { useDebounceFn } from "@vueuse/core";
import { ref, watch } from "vue";
import type { ComponentExposed } from "vue-component-type-helpers";
import PagedTable from "@/components/v2/Model/PagedTable.vue";
import { useProjectV1Store } from "@/store";
import { DEFAULT_PROJECT_NAME, type ComposedProject } from "@/types";
import ProjectV1Table from "./ProjectV1Table.vue";

const props = defineProps<{
  search: string;
  sessionKey: string;
  includeDefault: boolean;
}>();

const projectStore = useProjectV1Store();

const projectPagedTable =
  ref<ComponentExposed<typeof PagedTable<ComposedProject>>>();

const fetchProjects = async ({
  pageToken,
  pageSize,
}: {
  pageToken: string;
  pageSize: number;
}) => {
  const { nextPageToken, projects } = await projectStore.fetchProjectList({
    showDeleted: false,
    pageToken,
    pageSize,
    query: props.search,
  });
  return {
    nextPageToken: nextPageToken ?? "",
    list: projects.filter((project) => {
      if (!props.includeDefault && project.name === DEFAULT_PROJECT_NAME) {
        return false;
      }
      return true;
    }),
  };
};

watch(
  () => props.search,
  useDebounceFn(async () => {
    await projectPagedTable.value?.refresh();
  }, 500)
);
</script>

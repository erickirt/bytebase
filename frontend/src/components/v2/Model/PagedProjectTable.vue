<template>
  <PagedTable
    ref="projectPagedTable"
    :session-key="sessionKey"
    :fetch-list="fetchProjects"
  >
    <template #table="{ list, loading }">
      <ProjectV1Table
        v-bind="$attrs"
        :loading="loading"
        :project-list="list"
        :selected-project-names="selectedProjectNames"
        :show-selection="showSelection"
        :keyword="filter.query"
        @update:selected-project-names="updateSelectedProjectNames"
      />
    </template>
  </PagedTable>
</template>

<script lang="tsx" setup>
import { ref, watch } from "vue";
import type { ComponentExposed } from "vue-component-type-helpers";
import PagedTable from "@/components/v2/Model/PagedTable.vue";
import { useProjectV1Store, type ProjectFilter } from "@/store";
import { type ComposedProject } from "@/types";
import ProjectV1Table from "./ProjectV1Table.vue";

const props = withDefaults(
  defineProps<{
    filter: ProjectFilter;
    sessionKey: string;
    selectedProjectNames?: string[];
    showSelection?: boolean;
  }>(),
  {
    selectedProjectNames: () => [],
    showSelection: false,
  }
);

const emit = defineEmits<{
  (event: "update:selected-project-names", projectNames: string[]): void;
}>();

const projectStore = useProjectV1Store();

const projectPagedTable =
  ref<ComponentExposed<typeof PagedTable<ComposedProject>>>();

const updateSelectedProjectNames = (projectNames: string[]) => {
  emit("update:selected-project-names", projectNames);
};

const refresh = () => {
  projectPagedTable.value?.refresh();
};

defineExpose({ refresh });

const fetchProjects = async ({
  pageToken,
  pageSize,
}: {
  pageToken: string;
  pageSize: number;
}) => {
  const { nextPageToken, projects } = await projectStore.fetchProjectList({
    pageToken,
    pageSize,
    filter: props.filter,
  });
  return {
    nextPageToken: nextPageToken ?? "",
    list: projects,
  };
};

watch(
  () => props.filter,
  () => projectPagedTable.value?.refresh(),
  { deep: true }
);
</script>

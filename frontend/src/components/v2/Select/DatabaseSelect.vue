<template>
  <ResourceSelect
    v-bind="$attrs"
    class="bb-database-select"
    :remote="true"
    :loading="state.loading"
    :placeholder="$t('database.select')"
    :multiple="multiple"
    :value="databaseName"
    :values="databaseNames"
    :options="options"
    :custom-label="renderLabel"
    @search="handleSearch"
    @update:value="(val) => $emit('update:database-name', val)"
    @update:values="(val) => $emit('update:database-names', val)"
  />
</template>

<script lang="ts" setup>
import { useDebounceFn } from "@vueuse/core";
import { computed, h, watch, reactive, watchEffect } from "vue";
import { useI18n } from "vue-i18n";
import { useDatabaseV1Store } from "@/store";
import { workspaceNamePrefix } from "@/store/modules/v1/common";
import type { ComposedDatabase } from "@/types";
import {
  DEBOUNCE_SEARCH_DELAY,
  isValidDatabaseName,
  unknownDatabase,
} from "@/types";
import { type Engine } from "@/types/proto-es/v1/common_pb";
import {
  instanceV1Name,
  supportedEngineV1List,
  getDefaultPagination,
} from "@/utils";
import { InstanceV1EngineIcon } from "../Model";
import ResourceSelect from "./ResourceSelect.vue";

interface LocalState {
  loading: boolean;
  rawDatabaseList: ComposedDatabase[];
}

const props = withDefaults(
  defineProps<{
    databaseName?: string; // UNKNOWN_DATABASE_NAME stands for "ALL"
    databaseNames?: string[];
    environmentName?: string;
    projectName?: string;
    allowedEngineTypeList?: Engine[];
    includeAll?: boolean;
    autoReset?: boolean;
    filter?: (database: ComposedDatabase, index: number) => boolean;
    multiple?: boolean;
    clearable?: boolean;
    defaultSelectFirst?: boolean;
  }>(),
  {
    databaseName: undefined,
    databaseNames: undefined,
    environmentName: undefined,
    projectName: undefined,
    allowedEngineTypeList: () => supportedEngineV1List(),
    includeAll: false,
    autoReset: true,
    filter: undefined,
    multiple: false,
    clearable: false,
    defaultSelectFirst: false,
  }
);

const emit = defineEmits<{
  (event: "update:database-name", value: string | undefined): void;
  (event: "update:database-names", value: string[]): void;
}>();

const { t } = useI18n();
const databaseStore = useDatabaseV1Store();

const state = reactive<LocalState>({
  loading: true,
  rawDatabaseList: [],
});

const initSelectedDatabases = async (databaseNames: string[]) => {
  for (const databaseName of databaseNames) {
    if (isValidDatabaseName(databaseName)) {
      const db = await databaseStore.getOrFetchDatabaseByName(databaseName);
      if (!state.rawDatabaseList.find((d) => d.name === db.name)) {
        state.rawDatabaseList.unshift(db);
      }
    }
  }
};

const searchDatabases = async (name: string) => {
  const { databases } = await databaseStore.fetchDatabases({
    parent: props.projectName ?? `${workspaceNamePrefix}-`,
    filter: {
      environment: props.environmentName,
      engines: props.allowedEngineTypeList,
      query: name,
    },
    pageSize: getDefaultPagination(),
  });
  return databases;
};

const handleSearch = useDebounceFn(async (search: string) => {
  state.loading = true;
  try {
    const databases = await searchDatabases(search);
    state.rawDatabaseList = databases;
    if (!search) {
      if (props.includeAll) {
        const dummyAll = {
          ...unknownDatabase(),
          databaseName: t("database.all"),
        };
        state.rawDatabaseList.unshift(dummyAll);
      }
      if (props.databaseName) {
        await initSelectedDatabases([props.databaseName]);
      }
      if (props.databaseNames) {
        await initSelectedDatabases(props.databaseNames);
      }
    }
  } finally {
    state.loading = false;
  }
}, DEBOUNCE_SEARCH_DELAY);

watch(
  () => [props.environmentName, props.allowedEngineTypeList],
  () => {
    handleSearch("");
  },
  {
    deep: true,
    immediate: true,
  }
);

const combinedDatabaseList = computed(() => {
  if (props.filter) {
    return state.rawDatabaseList.filter(props.filter);
  }

  return state.rawDatabaseList;
});

const options = computed(() => {
  return combinedDatabaseList.value.map((database) => {
    return {
      resource: database,
      value: database.name,
      label: database.databaseName,
    };
  });
});

watchEffect(() => {
  if (!props.defaultSelectFirst || props.multiple) {
    return;
  }
  if (options.value.length === 0) {
    return;
  }

  emit("update:database-name", options.value[0].value);
});

const renderLabel = (database: ComposedDatabase) => {
  const children = [h("div", {}, [database.databaseName])];
  if (isValidDatabaseName(database.name)) {
    // prefix engine icon
    children.unshift(
      h(InstanceV1EngineIcon, {
        class: "mr-1",
        instance: database.instanceResource,
      })
    );
    // suffix engine name
    children.push(
      h(
        "div",
        {
          class: "text-xs opacity-60 ml-1",
        },
        [`(${instanceV1Name(database.instanceResource)})`]
      )
    );
  }
  return h(
    "div",
    {
      class: "w-full flex flex-row justify-start items-center truncate",
    },
    children
  );
};

// The database list might change if environment changes, and the previous selected id
// might not exist in the new list. In such case, we need to invalidate the selection
// and emit the event.
const resetInvalidSelection = () => {
  if (!props.autoReset || props.multiple) return;
  if (
    !state.loading &&
    props.databaseName &&
    !combinedDatabaseList.value.find((item) => item.name === props.databaseName)
  ) {
    emit("update:database-name", undefined);
  }
};

watch(() => combinedDatabaseList.value, resetInvalidSelection, {
  immediate: true,
  deep: true,
});
</script>

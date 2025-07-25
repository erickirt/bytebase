<template>
  <div
    class="flex h-full w-full flex-col justify-start items-stretch overflow-hidden bg-dark-bg"
  >
    <EditorAction @execute="handleExecute" @clear-screen="handleClearScreen" />

    <div
      v-if="!tabStore.isDisconnected"
      ref="queryListContainerRef"
      class="w-full flex-1 overflow-y-auto bg-dark-bg"
    >
      <div
        ref="queryListRef"
        class="w-full flex flex-col"
        :data-height="queryListHeight"
      >
        <div v-for="query in queryList" :key="query.id" class="relative">
          <Suspense>
            <CompactSQLEditor
              v-model:content="query.statement"
              class="min-h-[2rem]"
              :class="[
                isEditableQueryItem(query)
                  ? 'active-editor'
                  : 'read-only-editor',
              ]"
              :readonly="!isEditableQueryItem(query)"
              @execute="handleExecute"
              @history="handleHistory"
              @clear-screen="handleClearScreen"
            />
            <template #fallback>
              <div
                class="w-full min-h-[2rem] flex flex-col items-center justify-center"
              >
                <BBSpin />
              </div>
            </template>
          </Suspense>
          <ResultViewV1
            v-if="query.params && query.resultSet"
            class="flex-1 flex flex-col overflow-hidden max-h-80 p-2"
            :execute-params="query.params"
            :result-set="query.resultSet"
            :database="
              databaseStore.getDatabaseByName(query.params.connection.database)
            "
            :context-id="query.id"
            :loading="query.status === 'RUNNING'"
            :dark="true"
          />

          <div
            v-if="query.resultSet?.error"
            class="p-2 pb-1 text-md font-normal text-matrix-green-hover"
          >
            {{ $t("sql-editor.connection-lost") }}
          </div>

          <div
            v-if="query.status === 'RUNNING'"
            class="absolute inset-0 bg-black/20 flex justify-center items-center gap-2"
          >
            <BBSpin />
            <div
              v-if="query === currentQuery && expired"
              class="text-gray-400 cursor-pointer hover:underline text-sm select-none"
              @click="handleCancelQuery"
            >
              {{ $t("common.cancel") }}
            </div>
          </div>
        </div>
      </div>
    </div>
    <ConnectionHolder v-else />
  </div>
</template>

<script lang="ts" setup>
import { useElementSize } from "@vueuse/core";
import { computed, defineAsyncComponent, ref, unref, watch } from "vue";
import { watchEffect } from "vue";
import { BBSpin } from "@/bbkit";
import type { IStandaloneCodeEditor } from "@/components/MonacoEditor";
import {
  useSQLEditorTabStore,
  useDatabaseV1Store,
  useWebTerminalStore,
  batchGetOrFetchDatabases,
} from "@/store";
import type { SQLEditorQueryParams, WebTerminalQueryItemV1 } from "@/types";
import {
  EditorAction,
  ConnectionHolder,
  ResultViewV1,
} from "../../EditorCommon";
import { useHistory } from "./useHistory";

const CompactSQLEditor = defineAsyncComponent(
  () => import("./CompactSQLEditor.vue")
);

const tabStore = useSQLEditorTabStore();
const webTerminalStore = useWebTerminalStore();
const databaseStore = useDatabaseV1Store();

const queryState = computed(() => {
  return webTerminalStore.getQueryStateByTab(tabStore.currentTab!);
});

const queryList = computed(() => {
  return unref(queryState.value.queryItemList);
});

watchEffect(async () => {
  await batchGetOrFetchDatabases(
    queryList.value.map((query) => query?.params?.connection.database ?? "")
  );
});

const queryListContainerRef = ref<HTMLDivElement>();
const queryListRef = ref<HTMLDivElement>();

const currentQuery = computed(
  () => queryList.value[queryList.value.length - 1]
);

const { move: moveHistory } = useHistory();

const timer = computed(() => {
  return unref(queryState.value.timer);
});
const expired = computed(() => {
  return unref(timer.value.expired);
});

const isEditableQueryItem = (item: WebTerminalQueryItemV1): boolean => {
  return item === currentQuery.value && item.status === "IDLE";
};

const handleExecute = async (params: SQLEditorQueryParams) => {
  if (currentQuery.value.status !== "IDLE") {
    return;
  }

  // Prevent executing empty query;
  if (!params.statement) {
    console.warn("Empty query");
    return;
  }

  queryState.value.controller.events.emit("query", params);
};

const handleClearScreen = () => {
  const list = queryList.value;
  while (list.length > 1) {
    list.shift();
  }
};

const handleHistory = (
  direction: "up" | "down",
  editor: IStandaloneCodeEditor
) => {
  if (currentQuery.value.status !== "IDLE") {
    return;
  }
  moveHistory(direction);

  requestAnimationFrame(() => {
    const model = editor.getModel();
    if (!model) return;
    const lineNumber = model.getLineCount();
    const column = model.getLineMaxColumn(lineNumber);
    editor.setPosition({
      lineNumber,
      column,
    });
  });
};

const handleCancelQuery = async () => {
  queryState.value.controller.abort();
};

const { height: queryListHeight } = useElementSize(queryListRef);

watch(queryListHeight, () => {
  // Always scroll to the bottom as if we are in a real terminal.
  requestAnimationFrame(() => {
    const container = queryListContainerRef.value;
    if (container) {
      container.scrollTo(0, container.scrollHeight);
    }
  });
});
</script>

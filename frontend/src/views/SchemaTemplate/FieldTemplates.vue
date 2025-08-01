<template>
  <div class="w-full space-y-4 text-sm">
    <div v-if="!readonly" class="space-y-4">
      <div class="flex items-center justify-between gap-x-6">
        <div class="flex-1 textinfolabel">
          {{ $t("schema-template.field-template.description") }}
        </div>
      </div>
    </div>
    <div class="flex items-center justify-between gap-x-2 my-4">
      <div class="flex flex-row items-center justify-start gap-x-4">
        <template v-if="showEngineFilter">
          <NCheckbox
            v-for="item in engineList"
            :key="item"
            :checked="state.selectedEngine.has(item)"
            @update:checked="toggleEngineCheck(item)"
          >
            <div class="flex items-center gap-x-1 text-sm text-gray-600">
              <EngineIcon :engine="item" custom-class="ml-0 mr-1" />
              <span
                class="items-center text-xs px-2 py-0.5 rounded-full bg-gray-200 text-gray-800"
              >
                {{ countTemplateByEngine(item) }}
              </span>
            </div>
          </NCheckbox>
        </template>
      </div>
      <div class="flex items-center space-x-2">
        <div class="hidden md:block">
          <SearchBox
            v-model:value="state.searchText"
            :autofocus="true"
            :placeholder="$t('schema-template.search-by-name-or-comment')"
          />
        </div>
        <NButton
          type="primary"
          :disabled="readonly"
          @click="createSchemaTemplate"
        >
          <template #icon>
            <PlusIcon class="h-4 w-4" />
          </template>
          {{ $t("common.add") }}
        </NButton>
      </div>
    </div>
    <FieldTemplateView
      :engine="engine"
      :readonly="!!readonly"
      :template-list="filteredTemplateList"
      @view="editSchemaTemplate"
      @apply="$emit('apply', $event)"
    />
  </div>
  <Drawer v-model:show="state.showDrawer">
    <FieldTemplateForm
      :readonly="!!readonly"
      :create="!state.template.column?.name"
      :template="state.template"
      @dismiss="state.showDrawer = false"
    />
  </Drawer>
</template>

<script lang="ts" setup>
import { create } from "@bufbuild/protobuf";
import { PlusIcon } from "lucide-vue-next";
import { NButton, NCheckbox } from "naive-ui";
import { v1 as uuidv1 } from "uuid";
import { reactive, computed, onMounted } from "vue";
import { EngineIcon } from "@/components/Icon";
import FieldTemplateForm from "@/components/SchemaTemplate/FieldTemplateForm.vue";
import FieldTemplateView from "@/components/SchemaTemplate/FieldTemplateView.vue";
import { engineList } from "@/components/SchemaTemplate/utils";
import { Drawer, SearchBox } from "@/components/v2";
import { useSettingV1Store } from "@/store";
import { Engine } from "@/types/proto-es/v1/common_pb";
import { ColumnCatalogSchema } from "@/types/proto-es/v1/database_catalog_service_pb";
import { ColumnMetadataSchema } from "@/types/proto-es/v1/database_service_pb";
import type { SchemaTemplateSetting_FieldTemplate } from "@/types/proto-es/v1/setting_service_pb";
import {
  Setting_SettingName,
  SchemaTemplateSetting_FieldTemplateSchema,
} from "@/types/proto-es/v1/setting_service_pb";

interface LocalState {
  template: SchemaTemplateSetting_FieldTemplate;
  showDrawer: boolean;
  searchText: string;
  selectedEngine: Set<Engine>;
}

const props = defineProps<{
  engine?: Engine;
  readonly?: boolean;
  showEngineFilter?: boolean;
}>();

defineEmits<{
  (event: "apply", item: SchemaTemplateSetting_FieldTemplate): void;
}>();

const initialTemplate = (): SchemaTemplateSetting_FieldTemplate =>
  create(SchemaTemplateSetting_FieldTemplateSchema, {
    id: uuidv1(),
    engine: props.engine ?? Engine.MYSQL,
    category: "",
    column: create(ColumnMetadataSchema, {
      name: "",
      type: "",
      nullable: false,
      comment: "",
      position: 0,
      characterSet: "",
      collation: "",
    }),
    catalog: create(ColumnCatalogSchema, {}),
  });

const state = reactive<LocalState>({
  showDrawer: false,
  template: initialTemplate(),
  searchText: "",
  selectedEngine: new Set<Engine>(),
});

onMounted(() => {
  if (props.engine) {
    state.selectedEngine.add(props.engine);
  }
});

const createSchemaTemplate = () => {
  state.template = initialTemplate();
  state.showDrawer = true;
};

const editSchemaTemplate = (template: SchemaTemplateSetting_FieldTemplate) => {
  state.template = template;
  state.showDrawer = true;
};

const toggleEngineCheck = (engine: Engine) => {
  if (state.selectedEngine.has(engine)) {
    state.selectedEngine.delete(engine);
  } else {
    state.selectedEngine.add(engine);
  }
};

const settingStore = useSettingV1Store();

const schemaTemplateList = computed(() => {
  const setting = settingStore.getSettingByName(
    Setting_SettingName.SCHEMA_TEMPLATE
  );
  return setting?.value?.value?.case === "schemaTemplateSettingValue"
    ? (setting.value.value.value.fieldTemplates ?? [])
    : [];
});

const countTemplateByEngine = (engine: Engine) => {
  return schemaTemplateList.value.filter(
    (template) => template.engine === engine
  ).length;
};

const filteredTemplateList = computed(() => {
  if (state.selectedEngine.size === 0) {
    return schemaTemplateList.value.filter(filterTemplateByKeyword);
  }
  return schemaTemplateList.value.filter(
    (template) =>
      state.selectedEngine.has(template.engine) &&
      filterTemplateByKeyword(template)
  );
});

const filterTemplateByKeyword = (
  template: SchemaTemplateSetting_FieldTemplate
) => {
  const keyword = state.searchText.trim().toLowerCase();
  if (!keyword) return true;
  if (template.column?.name.toLowerCase().includes(keyword)) {
    return true;
  }
  if (template.column?.userComment.toLowerCase().includes(keyword)) {
    return true;
  }
  return false;
};
</script>

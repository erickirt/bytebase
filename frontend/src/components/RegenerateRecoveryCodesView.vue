<template>
  <RecoveryCodesView
    :recovery-codes="props.recoveryCodes"
    @download="state.recoveryCodesDownloaded = true"
  />
  <div class="flex flex-row justify-between items-center mb-8">
    <NButton @click="emit('close')">
      {{ $t("common.cancel") }}
    </NButton>
    <NButton
      type="primary"
      :disabled="!state.recoveryCodesDownloaded"
      @click="regenerateRecoveryCodes"
    >
      {{ $t("two-factor.setup-steps.recovery-codes-saved") }}
    </NButton>
  </div>
</template>

<script lang="ts" setup>
import { NButton } from "naive-ui";
import { onMounted, reactive } from "vue";
import { useI18n } from "vue-i18n";
import { pushNotification, useCurrentUserV1, useUserStore } from "@/store";
import { UpdateUserRequest } from "@/types/proto/v1/user_service";
import RecoveryCodesView from "./RecoveryCodesView.vue";

interface LocalState {
  recoveryCodesDownloaded: boolean;
}

const props = withDefaults(
  defineProps<{
    recoveryCodes: string[];
  }>(),
  {
    recoveryCodes: () => [],
  }
);

const emit = defineEmits(["close"]);

const { t } = useI18n();
const userStore = useUserStore();
const state = reactive<LocalState>({
  recoveryCodesDownloaded: false,
});
const currentUser = useCurrentUserV1();

onMounted(() => {
  regenerateTempMfaSecret();
});

const regenerateTempMfaSecret = async () => {
  await userStore.updateUser(
    UpdateUserRequest.fromPartial({
      user: {
        name: currentUser.value.name,
      },
      updateMask: [],
      regenerateTempMfaSecret: true,
    })
  );
};

const regenerateRecoveryCodes = async () => {
  await userStore.updateUser(
    UpdateUserRequest.fromPartial({
      user: {
        name: currentUser.value.name,
      },
      updateMask: [],
      regenerateRecoveryCodes: true,
    })
  );
  pushNotification({
    module: "bytebase",
    style: "SUCCESS",
    title: t("two-factor.messages.recovery-codes-regenerated"),
  });
  emit("close");
};
</script>

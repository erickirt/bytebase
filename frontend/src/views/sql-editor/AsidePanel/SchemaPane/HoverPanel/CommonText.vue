<template>
  <div class="max-w-[18rem] whitespace-pre-wrap break-all break-words">
    {{ content }}
  </div>
</template>

<script setup lang="ts">
import type { ComposedDatabase } from "@/types";
import type { DatabaseMetadata } from "@/types/proto/v1/database_service";

defineProps<{
  db: ComposedDatabase;
  database: DatabaseMetadata;
  content: string;
}>();
</script>

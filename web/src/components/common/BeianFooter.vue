<template>
  <div class="beian-records">
    <a
      class="beian-record-link"
      href="https://beian.miit.gov.cn"
      target="_blank"
      rel="noopener noreferrer"
    >
      豫ICP备2026007932号-1
    </a>
    <a
      class="beian-record-link beian-record-link--mps"
      href="https://beian.mps.gov.cn/#/query/webSearch?code=41010902002279"
      target="_blank"
      rel="noopener noreferrer"
    >
      <img
        :src="publicSecurityBadge"
        alt=""
        aria-hidden="true"
        class="beian-record-icon"
      >
      <span>豫公网安备41010902002279号</span>
    </a>
  </div>
</template>

<script setup lang="ts">
import publicSecurityBadge from '@/assets/beian-icon.png'
</script>

<style scoped>
.beian-records {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-wrap: wrap;
  column-gap: 16px;
  row-gap: 6px;
}

.beian-record-link {
  color: var(--text-secondary, #6b7280);
  font-size: 13px;
  line-height: 1.3;
  text-align: center;
  text-decoration: none;
  transition: color 0.2s ease;
}

.beian-record-link:hover {
  color: var(--accent, #0ea5e9);
}

.beian-record-link--mps {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.beian-record-icon {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
}
</style>

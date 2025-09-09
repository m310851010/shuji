import {main} from "@wailsjs/models";
import {GetEnv} from "@wailsjs/go";

/**
 * 环境变量
 */
export function useEnv() {
  const env = ref<main.EnvResult>(<main.EnvResult>{});
  onMounted(async () => {
    env.value = await GetEnv();
  })
  return env
}

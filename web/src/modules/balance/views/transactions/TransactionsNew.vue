<template>
  <div>
    <transaction-detail :transaction.sync="transaction" :is-edit="false" />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import { Routes } from '@/router/names';
import TransactionDetail from '@/modules/balance/components/transactions/TransactionDetail.vue';
import {
  ESplitOption,
  TTransactionNew,
} from '@/modules/balance/types/transaction';

@Component({
  components: { TransactionDetail },
})
export default class TransactionsNew extends Vue {
  transaction: TTransactionNew = {
    summary: '',
    changes: {},
    description: '',
    participants: [],
    payers: [],
    time: new Date(),
    value: 0,
    splitType: ESplitOption.Equal,
  };

  created() {
    this.$navigation.title = 'New Transaction';
    this.$navigation.parent = { name: Routes.BalanceTransactions };
  }
}
</script>

<style scoped lang="scss"></style>

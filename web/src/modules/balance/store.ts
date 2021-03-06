import { Action, Module, Mutation, VuexModule } from 'vuex-module-decorators';
import {
  ITransaction,
  TTransactionChanges,
} from '@/modules/balance/types/transaction';
import { Vue } from 'vue-property-decorator';
import { IUser } from '@/model/user';

@Module({
  name: 'balance',
  namespaced: true,
})
export default class BalanceStore extends VuexModule {
  transactions: ITransaction[] = [];
  users: IUser[] = [];
  userNames: { [_: string]: IUser } = {};
  userIds: { [_: string]: IUser } = {};

  @Action({ commit: 'setUsers', rawError: true })
  async fetchUsers() {
    const res = await Vue.$api.get<IUser[]>('/users');
    if (!res.success) throw new Error(res.message);
    return res.data!;
  }

  @Mutation
  async setUsers(users: IUser[]) {
    this.users = users;
    for (const idx in users) {
      this.userNames[users[idx].name] = users[idx];
      this.userIds[users[idx].id] = users[idx];
    }
  }

  @Action({ rawError: true })
  async calcChanges(tx: ITransaction): Promise<TTransactionChanges> {
    const changes: TTransactionChanges = {};
    for (const u of tx.payers) {
      if (!changes[u.id]) {
        changes[u.id] = { value: 0, name: u.name };
      }
      changes[u.id].value += u.value;
    }
    for (const u of tx.participants) {
      if (!changes[u.id]) changes[u.id] = { value: 0, name: u.name };
      changes[u.id].value -= u.value;
    }
    return changes;
  }

  @Action({ rawError: true })
  async fetchTransactions() {
    const res = await Vue.$api.get<ITransaction[]>('/balance/transactions');
    if (!res.success) return [];

    if (!this.users.length || !Object.keys(this.userIds).length) {
      await this.fetchUsers();
    }
    const transactions = await Promise.all(
      res.data!.map(async tx => ({
        ...tx,
        changes: await this.calcChanges(tx),
      }))
    );
    this.setTransactions({ transactions });
    return transactions;
  }

  @Mutation
  setTransactions(param: { transactions: ITransaction[] }) {
    this.transactions = param.transactions;
  }

  @Action({ rawError: true })
  async submitNewTransaction(param: { tx: ITransaction }) {
    const res = await Vue.$api.post<ITransaction>('/balance/transactions', {
      transaction: param.tx,
    });
    if (!res.success) throw new Error(res.message);

    return res.data!;
  }

  @Action({ rawError: true })
  async submitEditTransaction(param: { tx: ITransaction }) {
    const res = await Vue.$api.put<ITransaction>(
      '/balance/transactions/' + param.tx.id,
      {
        transaction: param.tx,
      }
    );
    if (!res.success) throw new Error(res.message);

    return res.data!;
  }

  @Action({ rawError: true })
  async removeTransaction(param: { id: string }) {
    const res = await Vue.$api.delete<Boolean>(
      '/balance/transactions/' + param.id
    );
    if (!res.success) throw new Error(res.message);
    return !!res.data;
  }

  @Action({ rawError: true })
  async fetchTransaction(param: { id: string }) {
    const res = await Vue.$api.get<ITransaction>(
      '/balance/transactions/' + param.id
    );
    if (!res.success || !res.data) throw new Error(res.message);
    res.data.changes = await this.calcChanges(res.data);
    return res.data;
  }
}

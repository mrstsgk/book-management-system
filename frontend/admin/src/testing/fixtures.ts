import type {
  BookListItemResponse,
  BookResponse,
  CatalogResponse,
} from '@/api/generated/api.schemas'

// 生成 faker（get<Op>ResponseMock）は欠けた項目を乱数で埋めるため、書影の有無などを
// 確かめるテストが不安定になる。テストで値を決めない項目は、ここの固定値で埋める。

/** 本の詳細のテスト用データを作る。 */
export function createBookFixture(
  override: Partial<BookResponse> = {},
): BookResponse {
  return {
    id: 1,
    isbn: '9784873118703',
    title: '本',
    authors: '',
    publisher: '',
    publishedOn: '',
    amazonUrl: null,
    coverUrl: null,
    coverSource: null,
    titleOverride: null,
    summary: '',
    tags: [],
    comment: '',
    rating: 1,
    version: 1,
    ...override,
  }
}

/** 本の一覧の 1 行のテスト用データを作る。 */
export function createBookListItemFixture(
  override: Partial<BookListItemResponse> = {},
): BookListItemResponse {
  return {
    id: 1,
    isbn: '9784873118703',
    title: '本',
    summary: '',
    tags: [],
    authors: '',
    amazonUrl: null,
    coverUrl: null,
    coverSource: null,
    rating: 1,
    ...override,
  }
}

/** 外部カタログの書誌のテスト用データを作る。 */
export function createCatalogFixture(
  override: Partial<CatalogResponse> = {},
): CatalogResponse {
  return {
    isbn: '9784873118703',
    title: '本',
    authors: '',
    publisher: '',
    publishedOn: '',
    amazonUrl: null,
    coverUrl: null,
    coverSource: null,
    ...override,
  }
}

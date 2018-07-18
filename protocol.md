## 2 Основные разделы

### 2.1 Дневник 

#### 2.1.2 Получение подробностей урока.

```json
{
    "action": "get_lesson_description",
    "id": "string|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id урока для которого мы хотим узнать подробности. |

Ответ на запрос:

```json
{
	"description": "string|undefined",
	"file": "string|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `description`    | Подробности задания. |
| `file`    | Ссылка (на наш сервер), по которой наш сервер будет стримить присоединенный файл с сайта. |


#### 2.1.3 Отметка задания как выполненного.

```json
{
    "action": "mark_as_done",
    "id": "number"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id домашнего задания, которое мы хотим отметить как выполненное. |

Ответ на запрос:

```json
{
	"success" : "boolean"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | Получилось ли выполнить запрос. |

#### 2.1.4 Отметка задания как невыполненного.

```json
{
    "action": "unmark_as_done",
    "id": "number"
}
```

Ответ на запрос:

```json
{
	"success": "boolean"
}
```

### 2.2 Объявления 

#### 2.2.1 Получение списка объявлений.

```json
{
    "action": "get_posts"
}
```


Ответ на запрос:

```json
{
	"posts": [{
		"id": "number",
		"unread": "boolean",
		"author": "string",
		"title": "string",
		"date": "string",
		"message": "string|undefined",
		"file": "string|undefined"
	}]
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id объявления. |
| `unread`    | `true` если объявление новое. `false` иначе. |
| `author`    | Автор объявления. |
| `title`    | Тема объявления. |
| `date`    | Дата. |
| `message`    | Сообщение. |
| `file`    | Ссылка (на наш сервер), по которой наш сервер будет стримить присоединенный файл с сайта. |

### 2.4 Отчеты

#### 2.4.7 Получение списка классов для отчета "Отчет о доступе к классному журналу".

```json
{
    "action": "get_report_journal_access_classes_list",
    "id": "number|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Если у нас аккаунт родителя, у которого несколько детей учится в школе, то ему на сайте в дневнике доступно между ними переключение. Это id выбранного ребенка. |

Ответ на запрос: 

```json
{
	"classes_list": [{
		"id": "number",
		"name": "string"
	}]
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id класса. |
| `name`    | Класс. |

#### 2.4.9 Получение данных для отчета "Информационное письмо для родителей".

```json
{
    "action": "get_report_parent_info_letter_data",
    "id": "number|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Если у нас аккаунт родителя, у которого несколько детей учится в школе, то ему на сайте в дневнике доступно между ними переключение. Это id выбранного ребенка. |

Ответ на запрос: 

```json
{
	"data": [{
		"report_types": [{
			"report_type_id": "number",
			"report_type_name": "string"
		}],
		"periods": [{
			"period_id": "number",
			"period_name": "string"
		}]
	}]
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `report_type_id`    | Id вида отчета. |
| `report_type_name`    | Название вида отчета. |
| `period_id`    | Id периода. |
| `period_name`    | Название периода. |

### 2.5 Школьные ресурсы

#### 2.5.1 Получения списка школьных файлов

```json
{
    "action": "get_recources"
}
```

Ответ на запрос:

```json
{
	"groups": [
		"group_title": "string",
		"files": [
			"name": "string",
			"link": "string"
		],
		"subgroups": [
			"subgroup_title": "string",
			"files": [
				"name": "string",
				"link": "string"
			]
		]
	]
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `group_title`    | Название группы. |
| `subgroup_title`    | Название подгруппы если имеется. |
| `name`    | Название файла. |
| `link`    | Ссылка на файл.|

### 2.6 Почта

#### 2.6.1 Получение списка писем

```json
{
    "action": "get_mail",
    "section": "number",
    "page": "number|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `page`    | Номер страницы, которую надо запросить у сайта. Undefined если первая страница. |
| `section`    | `0` если раздел `Входящие`,`1` если раздел `Черновики`,`2` если раздел `Отправленные`,`3` если раздел `Удаленные` |

Ответ на запрос:

```json
{	
	"letters": [
		"date": "string",
		"id": "number",
		"author": "string",
		"title": "string",
		"unread": "boolean"
	]
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `date`    | Дата отправки письма. |
| `id`    | Id письма. |
| `author`    | Отправитель. |
| `title`    | Тема. |
| `unread`    | `true` если письмо не прочитано,`false` иначе. |

#### 2.6.2 Получение подробностей письма

```json
{
    "action": "get_mail_description",
    "id": "number"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id письма. |

Ответ на запрос:

```json
{	
	"to": [
		"name": "string",
		"id": "number"
	],
	"copy": [
		"name": "string",
		"id": "number"
	],
	"description": "string",
	"files": [
		"file_name": "string",
		"link": "string"
	]
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `to`    | Массив пользователей, которым письмо адресовано. |
| `copy`    | Массив пользователей, которые находятся в копии письма. |
| `name`    | Имя, под которым пользователь записан в адресной книге. |
| `id`    | Id пользователя. |
| `description`    | Тело письма. |
| `files`    | Массив присоединенных файлов. |
| `file_name`    | Название файла. |
| `link`    | Ссылка (на наш сервер), по которой наш сервер будет стримить присоединенный файл с сайта. |

#### 2.6.3 Удаление письма

```json
{
    "action": "delete_mail",
    "id": "number",
    "section": "number"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id письма. |
| `section`    | `0` если раздел `Входящие`,`1` если раздел `Черновики`,`2` если раздел `Отправленные`,`3` если раздел `Удаленные` |

Ответ на запрос:

```json
{	
	"success": "boolean"
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | Получилось ли удалить письмо. |

#### 2.6.4 Отправка нового письма

```json
{
    "action": "send_letter",
    "to": [
		"id": "number"
	],
	"copy": [
		"id": "number"
	],
	"hidden_copy": [
		"id": "number"
	],
	"title": "string|undefined",
	"description": "string",
	"notification": "boolean"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `to`    | Массив пользователей, которым письмо адресовано. |
| `copy`    | Массив пользователей, которые находятся в копии письма. |
| `hidden_copy`    | Массив пользователей, которые находятся в скрытой копии письма. |
| `id`    | Id пользователя. |
| `title`    | Тема письма. |
| `description`    | Тело письма. |
| `notification`    | `true` если нужно уведомить отправителя о прочтении,`false` иначе. |

Ответ на запрос:

```json
{	
	"success": "boolean"
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | Получилось ли отправить письмо. |

#### 2.6.5 Загрузка адресной книги

```json
{
    "action": "get_adress_book"
}
```

Ответ на запрос:

```json
{	
	"adress_book": [
		"groups": [
			"title": "string",
			"users": [
				"name": "string",
				"id" : "number"
			]
		],
		"classes": [
			"class_name": "string",
			"users": [
				"student": "string",
				"student_id": "number",
				"parents": [
					"parent": "string",
					"parent_id": "number"
				]
			]
			"name": "string",
			"id" : "number"
		]
	]
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `title`    | Заголовок группы. |
| `name`    | Имя пользователя в адресной книге. |
| `id`    | Id пользователя в адресной книге. |
| `class_name`    | Название класса из адресной книги. |
| `student`    | Имя ученика в адресной книге. |
| `student_id`    | Id ученика в адресной книге. |
| `parents`    | Массив родителей ученика. Пустой если таковых нет. |
| `parent`    | Имя родителя в адресной книге. |
| `parent_id`    | Id родителя в адресной книге. |

### 2.7 Форум

#### 2.7.1 Загрузка списка тем 

```json
{
    "action": "get_forum",
    "page": "number|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `page`    | Номер страницы, которую надо запросить у сайта. Undefined если первая страница. |

Ответ на запрос:

```json
{	
	"posts": [
		"date": "string",
		"last_author": "string",
		"id": "number",
		"creator": "string",
		"answers": "number",
		"title": "string",
		"unread": "boolean"
	]
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `date`    | Дата последнего сообщения в теме. |
| `last_author`    | Автор последнего сообщения. |
| `id`    | Id темы. |
| `creator`    | Создатель темы. |
| `answers`    | Количество ответов. |
| `title`    | Тема. |
| `unread`    | `true` если пользователь не читал эту тему или тема содержит новые сообщения,`false` иначе. |

#### 2.7.2 Загрузка сообщений из темы

```json
{
    "action": "get_forum",
    "id": "number",
    "page": "number|undefined"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `page`    | Номер страницы, которую надо запросить у сайта. Undefined если первая страница. |
| `id`    | Id темы. |

Ответ на запрос:

```json
{	
	"messages": [
		"date": "string",
		"author": "string",
		"role": "string",
		"message": "string",
		"unread": "string"
	]
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `date`    | Дата сообщения. |
| `author`    | Автор сообщения. |
| `role`    | Роль пользователя, оставившего сообщение в системе. Примеры: `ученик`, `учитель`. |
| `message`    | Сообщение. |
| `unread`    | `true` если пользователь не читал это сообщение,`false` иначе. |

#### 2.7.3 Создание новой темы на форуме

```json
{
    "action": "create_topic",
    "title": "string",
    "message": "string"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `title`    | Название темы. |
| `message`    | Сообщение. |

Ответ на запрос:

```json
{	
	"success": "boolean"
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | `true` если тема успешно создана,`false` иначе. |

#### 2.7.4 Создание нового сообщения в теме

```json
{
    "action": "create_message_in_topic",
    "id": "number",
    "message": "string"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `id`    | Id темы. |
| `message`    | Сообщение. |

Ответ на запрос:

```json
{	
	"success": "boolean"
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | `true` если сообщение успешно отправлено,`false` иначе. |


### 2.8 Настройки

#### 2.8.1 Изменение пароля

```json
{
    "action": "change_password",
    "old": "string",
    "new": "string"
}
```

| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `old`    | Старый пароль. |
| `new`    | Новый пароль. |

Ответ на запрос:

```json
{	
	"success": "boolean"
}
```
| Ключ        | Значение                              |
| ----------- | ------------------------------------- |
| `success`    | `true` если пароль успешно изменен,`false` иначе. |


















from pydantic import BaseModel
from playhouse.kv import KeyValue


def set(key: str, obj: BaseModel, kv: KeyValue) -> None:
    """
    Set a key in the KeyValue store
    """
    kv[key] = obj.model_dump_json(by_alias=True)


def get[T](key: str, kv: KeyValue, model: type[T]) -> T:
    """
    Get a value from the KeyValue store
    """
    return model.model_validate_json(kv[key])


def get_all[T](kv: KeyValue, model: type[T]) -> list[T]:
    """
    Get all values from the KeyValue store
    """
    return [model.model_validate_json(val) for val in kv.values()]

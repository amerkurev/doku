from dataclasses import dataclass

from server.router.context import total_size


@dataclass
class Item:
    size: int


@dataclass
class ItemShared:
    shared_size: int


def test_total_size():
    # Test total_size function
    items = [
        Item(size=1),
        Item(size=2),
        Item(size=3),
    ]
    assert total_size(items) == '6 Bytes'

    items = [
        ItemShared(shared_size=1024),
        ItemShared(shared_size=1024 * 1024),
        ItemShared(shared_size=1024 * 1024 * 1024),
    ]
    assert total_size(items, field_name='shared_size') == '1.1 GB'

    assert total_size([]) == '0'

    assert total_size(None) == '0'
